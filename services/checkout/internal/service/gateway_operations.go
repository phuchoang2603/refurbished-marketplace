package service

import (
	"context"
	"database/sql"
	"errors"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/phuchoang2603/refurbished-marketplace/services/checkout/internal/database"
	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
	sharedlog "github.com/phuchoang2603/refurbished-marketplace/shared/observe/log"
	sharedtrace "github.com/phuchoang2603/refurbished-marketplace/shared/observe/trace"
	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
	"go.opentelemetry.io/otel/attribute"
)

func (service *Service) RunGatewayOperations(ctx context.Context) error {
	return runWorker(ctx, service.ProcessGatewayOperation, "gateway operation processing failed")
}

func (service *Service) ProcessGatewayOperation(ctx context.Context) (bool, error) {
	operation, err := database.New(service.db).ClaimGatewayOperation(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	headers := map[string]string{}
	for _, line := range strings.Split(operation.Tracingspancontext.String, "\n") {
		if key, value, ok := strings.Cut(line, "="); ok {
			headers[key] = value
		}
	}
	ctx = sharedtrace.ContextFromHeaders(ctx, headers)
	ctx, span := sharedtrace.Tracer("checkout").Start(ctx, "gateway.create_payment")
	defer span.End()
	row, err := database.New(service.db).GetCheckout(ctx, operation.CheckoutID)
	if err != nil {
		return false, err
	}
	span.SetAttributes(attribute.String("order_id", row.OrderID.UUID.String()))
	if row.State != checkoutv1.CheckoutState_CHECKOUT_STATE_CREATING_PAYMENT.String() {
		return true, database.New(service.db).FinishGatewayOperation(ctx, database.FinishGatewayOperationParams{CheckoutID: row.ID, Status: "failed"})
	}
	var payment gatewayPayment
	var callErr error
	if row.DeadlineAt.Valid && time.Now().Before(row.DeadlineAt.Time) {
		callCtx, cancel := context.WithDeadline(ctx, row.DeadlineAt.Time)
		payment, callErr = service.gateway.create(callCtx, row.OrderID.UUID.String(), operation.Request)
		cancel()
	} else {
		callErr = errGatewayDefect
	}
	tx, err := service.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback() }()
	queries := database.New(tx)
	row, err = queries.LockCheckout(ctx, row.ID)
	if err != nil {
		return false, err
	}
	if row.State != checkoutv1.CheckoutState_CHECKOUT_STATE_CREATING_PAYMENT.String() {
		if err := queries.FinishGatewayOperation(ctx, database.FinishGatewayOperationParams{CheckoutID: row.ID, Status: "failed"}); err != nil {
			return false, err
		}
		return true, tx.Commit()
	}
	deadlineExceeded := !row.DeadlineAt.Valid || !time.Now().Before(row.DeadlineAt.Time)
	if errors.Is(callErr, errGatewayUnavailable) && !deadlineExceeded {
		delay := time.Second * time.Duration(1<<min(operation.Attempts, 5))
		delay = time.Duration(float64(delay) * (0.75 + rand.Float64()*0.5))
		next := time.Now().Add(delay)
		if next.After(row.DeadlineAt.Time) {
			next = row.DeadlineAt.Time
		}
		if err := queries.RetryGatewayOperation(ctx, database.RetryGatewayOperationParams{CheckoutID: row.ID, Attempts: operation.Attempts, NextAttemptAt: next}); err != nil {
			return false, err
		}
		sharedlog.InfoContext(ctx, "gateway creation will retry", "order_id", row.OrderID.UUID.String(), "attempt", operation.Attempts)
		return true, tx.Commit()
	}
	if callErr != nil || deadlineExceeded {
		if err := queries.FinishGatewayOperation(ctx, database.FinishGatewayOperationParams{CheckoutID: row.ID, Status: "failed"}); err != nil {
			return false, err
		}
		err = deadlineCommand(ctx, queries, row, checkoutv1.CheckoutState_CHECKOUT_STATE_RELEASING_STOCK, messaging.EventTypeCheckoutCancelStock, "cancel-stock", &checkoutv1.CancelStock{Reason: "payment preparation failed"}, time.Now().Add(preparationDeadline))
	} else {
		span.SetAttributes(attribute.String("gateway_payment_id", payment.PaymentID))
		err = queries.RecordGatewayPayment(ctx, database.RecordGatewayPaymentParams{ID: row.ID, GatewayPaymentID: sql.NullString{String: payment.PaymentID, Valid: true}, HostedCheckoutUrl: sql.NullString{String: payment.CheckoutURL, Valid: true}, GatewayExpiresAt: sql.NullTime{Time: payment.ExpiresAt, Valid: true}})
		if err == nil {
			_, err = queries.AdvanceCheckout(ctx, database.AdvanceCheckoutParams{ID: row.ID, Version: row.Version, State: checkoutv1.CheckoutState_CHECKOUT_STATE_READY_TO_PAY.String(), DeadlineAt: sql.NullTime{Time: payment.ExpiresAt.Add(service.config.PaymentDeadlineGrace), Valid: true}})
		}
		if err == nil {
			err = queries.FinishGatewayOperation(ctx, database.FinishGatewayOperationParams{CheckoutID: row.ID, Status: "succeeded"})
		}
		if err == nil {
			sharedlog.InfoContext(ctx, "gateway payment ready", "order_id", row.OrderID.UUID.String(), "gateway_payment_id", payment.PaymentID)
		}
	}
	if err != nil {
		return false, err
	}
	return true, tx.Commit()
}

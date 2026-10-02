package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/phuchoang2603/refurbished-marketplace/services/checkout/internal/database"
	"github.com/phuchoang2603/refurbished-marketplace/services/checkout/internal/grpcserver"
	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
	sharedtrace "github.com/phuchoang2603/refurbished-marketplace/shared/observe/trace"
	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const orderDeadline = 2 * time.Minute

func (service *Service) SubmitCheckout(ctx context.Context, request *checkoutv1.SubmitCheckoutRequest) (*checkoutv1.SubmitCheckoutResponse, error) {
	buyerID, merchantID, err := validateSubmit(request)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	canonical, err := (proto.MarshalOptions{Deterministic: true}).Marshal(request)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid checkout request")
	}
	hash := sha256.Sum256(canonical)
	checkoutID := uuid.New()
	orderID := uuid.New()
	snapshotRequest := proto.Clone(request).(*checkoutv1.SubmitCheckoutRequest)
	snapshotRequest.ReturnUrl = strings.ReplaceAll(request.GetReturnUrl(), "{order_id}", orderID.String())
	snapshot, err := protojson.Marshal(snapshotRequest)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid checkout snapshot")
	}
	operationID := uuid.NewSHA1(checkoutID, []byte("create-order"))
	command := &checkoutv1.CheckoutMessage{
		Context: &checkoutv1.MessageContext{
			MessageId: operationID.String(), CheckoutId: checkoutID.String(),
			OperationId: operationID.String(), OrderId: orderID.String(), WorkflowVersion: 1,
		},
		Payload: &checkoutv1.CheckoutMessage_CreateOrder{CreateOrder: &checkoutv1.CreateOrder{
			BuyerUserId: buyerID.String(), MerchantId: merchantID.String(),
			TotalCents: request.GetTotalCents(), Currency: request.GetCurrency(), Items: request.GetItems(),
		}},
	}
	payload, err := proto.Marshal(command)
	if err != nil {
		return nil, err
	}

	tx, err := service.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	queries := database.New(tx)
	row, err := queries.InsertCheckout(ctx, database.InsertCheckoutParams{
		ID: checkoutID, BuyerUserID: buyerID, MerchantID: merchantID,
		IntentKey: request.GetIntentKey(), RequestHash: hash[:], Snapshot: snapshot,
		State:      checkoutv1.CheckoutState_CHECKOUT_STATE_CREATING_ORDER.String(),
		OrderID:    uuid.NullUUID{UUID: orderID, Valid: true},
		DeadlineAt: sql.NullTime{Time: time.Now().Add(orderDeadline), Valid: true},
	})
	if errors.Is(err, sql.ErrNoRows) {
		row, err = queries.GetCheckoutByIntent(ctx, database.GetCheckoutByIntentParams{
			BuyerUserID: buyerID, IntentKey: request.GetIntentKey(),
		})
		if err != nil {
			return nil, err
		}
		if !bytes.Equal(row.RequestHash, hash[:]) {
			return nil, status.Error(codes.AlreadyExists, "checkout intent conflicts with original request")
		}
	} else if err != nil {
		return nil, err
	} else {
		if err := queries.InsertOutbox(ctx, database.InsertOutboxParams{
			ID: operationID, AggregateID: row.ID, EventType: messaging.EventTypeCheckoutCreateOrder,
			Payload: payload, Tracingspancontext: sql.NullString{String: sharedtrace.SerializeContext(ctx), Valid: true},
		}); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &checkoutv1.SubmitCheckoutResponse{CheckoutId: row.ID.String(), State: checkoutv1.CheckoutState_CHECKOUT_STATE_CREATING_ORDER}, nil
}

func (service *Service) GetCheckout(ctx context.Context, checkoutID uuid.UUID) (grpcserver.CheckoutView, error) {
	row, err := database.New(service.db).GetCheckout(ctx, checkoutID)
	if errors.Is(err, sql.ErrNoRows) {
		return grpcserver.CheckoutView{}, status.Error(codes.NotFound, "checkout not found")
	}
	if err != nil {
		return grpcserver.CheckoutView{}, err
	}
	state := checkoutv1.CheckoutState(checkoutv1.CheckoutState_value[row.State])
	view := grpcserver.CheckoutView{
		CheckoutID: row.ID, BuyerUserID: row.BuyerUserID,
		OrderID: row.OrderID.UUID.String(), State: state,
		FailureReason: row.FailureReason.String,
	}
	if state == checkoutv1.CheckoutState_CHECKOUT_STATE_READY_TO_PAY {
		view.PaymentSessionID = row.PaymentSessionID.String
		view.ReturnURL = row.PaymentReturnUrl.String
	}
	return view, nil
}

func validateSubmit(request *checkoutv1.SubmitCheckoutRequest) (uuid.UUID, uuid.UUID, error) {
	if request == nil {
		return uuid.Nil, uuid.Nil, errors.New("checkout request required")
	}
	buyerID, buyerError := uuid.Parse(request.GetBuyerUserId())
	merchantID, merchantError := uuid.Parse(request.GetMerchantId())
	if buyerError != nil || merchantError != nil || buyerID == uuid.Nil || merchantID == uuid.Nil {
		return uuid.Nil, uuid.Nil, errors.New("buyer and merchant IDs must be valid")
	}
	if strings.TrimSpace(request.GetIntentKey()) == "" || strings.TrimSpace(request.GetBuyerEmail()) == "" ||
		strings.TrimSpace(request.GetCurrency()) == "" || strings.TrimSpace(request.GetReturnUrl()) == "" ||
		request.GetShippingAddress() == nil || len(request.GetItems()) == 0 || request.GetTotalCents() <= 0 {
		return uuid.Nil, uuid.Nil, errors.New("checkout facts must be complete")
	}
	var total int64
	for _, item := range request.GetItems() {
		if item == nil || item.GetName() == "" || item.GetQuantity() <= 0 || item.GetUnitPriceCents() <= 0 {
			return uuid.Nil, uuid.Nil, errors.New("invalid checkout item")
		}
		productID, err := uuid.Parse(item.GetProductId())
		if err != nil || productID == uuid.Nil || item.GetUnitPriceCents() > (math.MaxInt64-total)/int64(item.GetQuantity()) {
			return uuid.Nil, uuid.Nil, errors.New("invalid checkout item or total")
		}
		total += item.GetUnitPriceCents() * int64(item.GetQuantity())
	}
	if total != request.GetTotalCents() {
		return uuid.Nil, uuid.Nil, errors.New("checkout total does not match items")
	}
	return buyerID, merchantID, nil
}

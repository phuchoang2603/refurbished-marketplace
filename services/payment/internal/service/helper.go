package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"refurbished-marketplace/services/payment/internal/database"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

func mapDBPaymentTransactionView(tx database.PaymentTransaction) PaymentTransactionView {
	v := PaymentTransactionView{
		ID:             tx.ID.String(),
		OrderID:        tx.OrderID.String(),
		MerchantID:     tx.MerchantID.String(),
		AmountCents:    tx.AmountCents,
		Currency:       tx.Currency,
		Status:         tx.Status,
		IdempotencyKey: tx.IdempotencyKey,
		CreatedAt:      tx.CreatedAt,
		UpdatedAt:      tx.UpdatedAt,
	}
	if tx.GatewayTransactionID.Valid {
		v.GatewayTransactionID = tx.GatewayTransactionID.String
	}
	return v
}

func mapDBHostedPaymentSessionView(intent database.PaymentIntent) HostedPaymentSessionView {
	v := HostedPaymentSessionView{
		OrderID:   intent.OrderID.String(),
		Currency:  intent.Currency,
		Status:    intent.Status,
		ReturnURL: intent.ReturnUrl,
		CancelURL: intent.CancelUrl,
		CreatedAt: intent.CreatedAt,
		UpdatedAt: intent.UpdatedAt,
		FailureReason: func() string {
			if intent.FailureReason.Valid {
				return intent.FailureReason.String
			}
			return ""
		}(),
	}
	if intent.PaymentSessionID.Valid {
		v.PaymentSessionID = intent.PaymentSessionID.String
	}
	if intent.ExpiresAt.Valid {
		v.ExpiresAt = intent.ExpiresAt.Time
	}
	return v
}

func loadPaymentTransaction(ctx context.Context, q *database.Queries, id uuid.UUID) (database.PaymentTransaction, error) {
	row, err := q.GetPaymentTransactionByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return database.PaymentTransaction{}, ErrTransactionNotFound
		}
		return database.PaymentTransaction{}, err
	}
	return row, nil
}

func loadPaymentIntentByOrderID(ctx context.Context, q *database.Queries, orderID uuid.UUID) (database.PaymentIntent, error) {
	row, err := q.GetPaymentIntentByOrderID(ctx, orderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return database.PaymentIntent{}, ErrIntentNotFound
		}
		return database.PaymentIntent{}, err
	}
	return row, nil
}

func parseOrderUUIDs(msg interface {
	GetOrderId() string
	GetMerchantId() string
},
) (orderID, merchantID uuid.UUID, err error) {
	orderID, err = uuid.Parse(msg.GetOrderId())
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("order_id: %w", err)
	}
	merchantID, err = uuid.Parse(msg.GetMerchantId())
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("merchant_id: %w", err)
	}
	return orderID, merchantID, nil
}

func optionalNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func optionalNullTime(t time.Time) sql.NullTime {
	return sql.NullTime{Time: t, Valid: !t.IsZero()}
}

func paymentTransactionIsTerminal(status string) bool {
	return status == PaymentTxStatusSucceeded || status == PaymentTxStatusFailed
}

func hostedPaymentSessionIsTerminal(status string) bool {
	return status == HostedPaymentSessionStatusSucceeded || status == HostedPaymentSessionStatusFailed || status == HostedPaymentSessionStatusCancelled || status == HostedPaymentSessionStatusExpired
}

func hostedPaymentSessionMapsToSuccess(status string) bool {
	return status == HostedPaymentSessionStatusSucceeded
}

func isPostgresUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}

func defaultPaymentCurrency(currency string) string {
	if currency == "" {
		return "USD"
	}
	return currency
}

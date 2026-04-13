package service

import (
	"encoding/json"
	"time"

	"refurbished-marketplace/services/payment/internal/database"

	"github.com/google/uuid"
)

// PaymentTransactionView is a transport-neutral snapshot for gRPC mapping.
type PaymentTransactionView struct {
	ID                   string
	OrderID              string
	OrderItemID          string
	MerchantID           string
	AmountCents          int64
	Currency             string
	Status               string
	IdempotencyKey       string
	GatewayTransactionID string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func paymentTransactionViewFromDB(tx database.PaymentTransaction) PaymentTransactionView {
	v := PaymentTransactionView{
		ID:             tx.ID.String(),
		OrderID:        tx.OrderID.String(),
		OrderItemID:    tx.OrderItemID.String(),
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

// InitiatePaymentParams is service-layer input (no protobuf).
type InitiatePaymentParams struct {
	OrderID         uuid.UUID
	BuyerUserID     uuid.UUID
	PaymentToken    string
	Currency        string
	BillingAddress  json.RawMessage
	ShippingAddress json.RawMessage
}


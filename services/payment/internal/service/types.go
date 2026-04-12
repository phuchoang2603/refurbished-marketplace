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

// OrderItemCreatedPayload matches the JSON emitted on the orders.item.created topic.
type OrderItemCreatedPayload struct {
	OrderID        string `json:"order_id"`
	OrderItemID    string `json:"order_item_id"`
	BuyerUserID    string `json:"buyer_user_id"`
	ProductID      string `json:"product_id"`
	MerchantID     string `json:"merchant_id"`
	Quantity       int32  `json:"quantity"`
	UnitPriceCents int64  `json:"unit_price_cents"`
	LineTotalCents int64  `json:"line_total_cents"`
}

type paymentItemOutboxPayload struct {
	OrderID     string `json:"order_id"`
	OrderItemID string `json:"order_item_id"`
}

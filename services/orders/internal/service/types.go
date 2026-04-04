package service

import (
	"time"

	"github.com/google/uuid"
)

type OrderItemInput struct {
	ProductID      uuid.UUID
	Quantity       int32
	UnitPriceCents int64
}

type Order struct {
	ID          uuid.UUID
	BuyerUserID uuid.UUID
	Status      string
	TotalCents  int64
	Items       []OrderItem
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type OrderItem struct {
	ID             uuid.UUID
	OrderID        uuid.UUID
	ProductID      uuid.UUID
	Quantity       int32
	UnitPriceCents int64
	LineTotalCents int64
	CreatedAt      time.Time
}

type outboxItem struct {
	ProductID      string `json:"product_id"`
	Quantity       int32  `json:"quantity"`
	UnitPriceCents int64  `json:"unit_price_cents"`
}
type outboxPayload struct {
	OrderID     string       `json:"order_id"`
	BuyerUserID string       `json:"buyer_user_id"`
	TotalCents  int64        `json:"total_cents"`
	Items       []outboxItem `json:"items"`
}

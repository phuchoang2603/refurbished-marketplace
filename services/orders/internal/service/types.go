package service

import (
	"time"

	"github.com/google/uuid"
)

type OrderItemInput struct {
	ProductID      uuid.UUID
	MerchantID     uuid.UUID
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
	MerchantID     uuid.UUID
	Quantity       int32
	UnitPriceCents int64
	LineTotalCents int64
	CreatedAt      time.Time
}

type outboxItem struct {
	OrderID        string `json:"order_id"`
	OrderItemID    string `json:"order_item_id"`
	BuyerUserID    string `json:"buyer_user_id"`
	ProductID      string `json:"product_id"`
	MerchantID     string `json:"merchant_id"`
	Quantity       int32  `json:"quantity"`
	UnitPriceCents int64  `json:"unit_price_cents"`
	LineTotalCents int64  `json:"line_total_cents"`
}

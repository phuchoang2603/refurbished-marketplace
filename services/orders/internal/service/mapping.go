package service

import (
	"time"

	"refurbished-marketplace/services/orders/internal/database"

	"github.com/google/uuid"
)

func mapOrderFields(id, buyerUserID, merchantID uuid.UUID, status string, totalCents int64, createdAt, updatedAt time.Time) Order {
	return Order{
		ID:          id,
		BuyerUserID: buyerUserID,
		MerchantID:  merchantID,
		Status:      status,
		TotalCents:  totalCents,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}

func mapDBOrderItem(i database.OrderItem) OrderItem {
	return OrderItem{
		ID:             i.ID,
		OrderID:        i.OrderID,
		ProductID:      i.ProductID,
		Quantity:       i.Quantity,
		UnitPriceCents: i.UnitPriceCents,
		LineTotalCents: i.LineTotalCents,
		CreatedAt:      i.CreatedAt,
	}
}

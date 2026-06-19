package service

import (
	"refurbished-marketplace/services/orders/internal/database"
)

func mapDBOrder(i database.Order) Order {
	return Order{
		ID:          i.ID,
		BuyerUserID: i.BuyerUserID,
		MerchantID:  i.MerchantID,
		Status:      i.Status,
		TotalCents:  i.TotalCents,
		CreatedAt:   i.CreatedAt,
		UpdatedAt:   i.UpdatedAt,
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

package service

import (
	"strings"

	"refurbished-marketplace/services/orders/internal/database"

	"github.com/google/uuid"
)

func mapDBOrder(o database.Order) Order {
	return Order{
		ID:          o.ID,
		BuyerUserID: o.BuyerUserID,
		Status:      o.Status,
		TotalCents:  o.TotalCents,
		CreatedAt:   o.CreatedAt,
		UpdatedAt:   o.UpdatedAt,
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

func validateCreateOrderInput(buyerUserID uuid.UUID, items []OrderItemInput, totalCents int64) error {
	if buyerUserID == uuid.Nil {
		return ErrInvalidBuyerID
	}
	if len(items) == 0 {
		return ErrInvalidProductID
	}
	if totalCents <= 0 {
		return ErrInvalidTotalCents
	}
	for _, item := range items {
		if item.ProductID == uuid.Nil {
			return ErrInvalidProductID
		}
		if item.Quantity <= 0 {
			return ErrInvalidQuantity
		}
		if item.UnitPriceCents <= 0 {
			return ErrInvalidUnitPriceCents
		}
	}
	return nil
}

func validateOrderStatus(status string) (string, error) {
	status = strings.TrimSpace(strings.ToUpper(status))
	if status == "" || status == OrderStatusUnspecified {
		return "", ErrInvalidStatus
	}
	if status != OrderStatusPending && status != OrderStatusPaid && status != OrderStatusFailed {
		return "", ErrInvalidStatus
	}
	return status, nil
}

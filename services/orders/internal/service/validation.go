package service

import (
	"strings"

	"github.com/google/uuid"
)

func validateCreateOrderInput(buyerUserID, merchantID uuid.UUID, items []OrderItemInput, totalCents int64) error {
	if buyerUserID == uuid.Nil {
		return ErrInvalidBuyerID
	}
	if merchantID == uuid.Nil {
		return ErrInvalidMerchantID
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

func validateListPagination(limit, offset int32) error {
	if limit <= 0 || limit > 100 {
		return ErrInvalidPagination
	}
	if offset < 0 {
		return ErrInvalidPagination
	}
	return nil
}

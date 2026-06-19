package service

import (
	"context"

	"github.com/google/uuid"
)

func (s *Service) GetPaymentTransaction(ctx context.Context, id uuid.UUID) (PaymentTransactionView, error) {
	row, err := loadPaymentTransaction(ctx, s.queries, id)
	if err != nil {
		return PaymentTransactionView{}, err
	}
	return mapDBPaymentTransactionView(row), nil
}

package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"refurbished-marketplace/services/payment/internal/database"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// loadPaymentTransaction loads a row or returns ErrTransactionNotFound (never sql.ErrNoRows).
func loadPaymentTransaction(ctx context.Context, q queryStore, id uuid.UUID) (database.PaymentTransaction, error) {
	row, err := q.GetPaymentTransactionByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return database.PaymentTransaction{}, ErrTransactionNotFound
		}
		return database.PaymentTransaction{}, err
	}
	return row, nil
}

// loadPaymentIntentByOrderID loads a row or returns ErrIntentNotFound (never sql.ErrNoRows).
func loadPaymentIntentByOrderID(ctx context.Context, q queryStore, orderID uuid.UUID) (database.PaymentIntent, error) {
	row, err := q.GetPaymentIntentByOrderID(ctx, orderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return database.PaymentIntent{}, ErrIntentNotFound
		}
		return database.PaymentIntent{}, err
	}
	return row, nil
}

// paymentTransactionExistsForOrderItem reports whether a transaction row exists for the order item.
func paymentTransactionExistsForOrderItem(ctx context.Context, q queryStore, orderItemID uuid.UUID) (exists bool, err error) {
	_, err = q.GetPaymentTransactionByOrderItemID(ctx, orderItemID)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return false, err
}

func parseOrderItemCreatedUUIDs(msg interface {
	GetOrderId() string
	GetOrderItemId() string
	GetMerchantId() string
}) (orderID, orderItemID, merchantID uuid.UUID, err error) {
	orderID, err = uuid.Parse(msg.GetOrderId())
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("order_id: %w", err)
	}
	orderItemID, err = uuid.Parse(msg.GetOrderItemId())
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("order_item_id: %w", err)
	}
	merchantID, err = uuid.Parse(msg.GetMerchantId())
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("merchant_id: %w", err)
	}
	return orderID, orderItemID, merchantID, nil
}

func optionalNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func paymentTransactionIsTerminal(status string) bool {
	return status == PaymentTxStatusSucceeded || status == PaymentTxStatusFailed
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

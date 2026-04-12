// Package service implements payment domain logic: intents, transactions, inbox/outbox, and Kafka handlers.
package service

import (
	"context"
	"database/sql"
	"errors"
	"refurbished-marketplace/services/payment/internal/database"

	"github.com/google/uuid"
)

type queryStore interface {
	UpsertPaymentIntent(ctx context.Context, arg database.UpsertPaymentIntentParams) (database.PaymentIntent, error)
	GetPaymentTransactionByID(ctx context.Context, id uuid.UUID) (database.PaymentTransaction, error)
	UpdatePaymentTransactionGatewayResult(ctx context.Context, arg database.UpdatePaymentTransactionGatewayResultParams) (database.PaymentTransaction, error)
	CreatePaymentOutbox(ctx context.Context, arg database.CreatePaymentOutboxParams) (database.PaymentOutbox, error)
	InsertPaymentInboxMessage(ctx context.Context, messageID string) error
	GetPaymentIntentByOrderID(ctx context.Context, orderID uuid.UUID) (database.PaymentIntent, error)
	GetPaymentTransactionByOrderItemID(ctx context.Context, orderItemID uuid.UUID) (database.PaymentTransaction, error)
	CreatePaymentTransaction(ctx context.Context, arg database.CreatePaymentTransactionParams) (database.PaymentTransaction, error)
}

var (
	ErrIntentNotFound      = errors.New("payment intent not found")
	ErrTransactionNotFound = errors.New("payment transaction not found")
)

const (
	PaymentTxStatusInitialized = "INITIALIZED"
	PaymentTxStatusSubmitted   = "SUBMITTED"
	PaymentTxStatusSucceeded   = "SUCCEEDED"
	PaymentTxStatusFailed      = "FAILED"
)

type Service struct {
	db      *sql.DB
	queries queryStore
}

func New(queries queryStore, db *sql.DB) *Service {
	return &Service{queries: queries, db: db}
}

func (s *Service) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

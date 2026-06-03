package service

import (
	"database/sql"
	"errors"

	"refurbished-marketplace/services/payment/internal/database"
)

var (
	ErrIntentNotFound      = errors.New("payment intent not found")
	ErrTransactionNotFound = errors.New("payment transaction not found")
	ErrSessionMismatch     = errors.New("payment session does not match order")
)

const (
	PaymentTxStatusInitialized = "INITIALIZED"
	PaymentTxStatusSubmitted   = "SUBMITTED"
	PaymentTxStatusSucceeded   = "SUCCEEDED"
	PaymentTxStatusFailed      = "FAILED"

	HostedPaymentSessionStatusPending   = "PENDING"
	HostedPaymentSessionStatusSucceeded = "SUCCEEDED"
	HostedPaymentSessionStatusFailed    = "FAILED"
	HostedPaymentSessionStatusCancelled = "CANCELLED"
	HostedPaymentSessionStatusExpired   = "EXPIRED"
)

type Service struct {
	db      *sql.DB
	queries *database.Queries
}

func New(db *sql.DB) *Service {
	return &Service{queries: database.New(db), db: db}
}

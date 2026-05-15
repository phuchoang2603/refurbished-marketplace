package service

import (
	"database/sql"
	"errors"

	"refurbished-marketplace/services/payment/internal/database"
)

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
	queries *database.Queries
}

func New(db *sql.DB) *Service {
	return &Service{queries: database.New(db), db: db}
}

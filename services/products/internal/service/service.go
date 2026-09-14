package service

import (
	"database/sql"
	"errors"

	"github.com/phuchoang2603/refurbished-marketplace/services/products/internal/database"
)

var (
	ErrInvalidInitialStock = errors.New("explicit non-negative initial stock is required")
	ErrInvalidProductName  = errors.New("invalid product name")
	ErrInvalidPrice        = errors.New("invalid product price")
	ErrInvalidMerchantID   = errors.New("invalid merchant id")
	ErrProductNotFound     = errors.New("product not found")
	ErrInvalidListLimit    = errors.New("invalid list limit")
	ErrInvalidListOffset   = errors.New("invalid list offset")
	ErrInvalidProductID    = errors.New("invalid product id")
	ErrInvalidBatchSize    = errors.New("invalid product batch size")
)

type Service struct {
	db      *sql.DB
	queries *database.Queries
}

func New(db *sql.DB) *Service {
	return &Service{db: db, queries: database.New(db)}
}

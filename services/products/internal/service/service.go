package service

import (
	"errors"

	"github.com/phuchoang2603/refurbished-marketplace/services/products/internal/catalog"
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
	store *catalog.Store
}

func New(store *catalog.Store) *Service {
	return &Service{store: store}
}

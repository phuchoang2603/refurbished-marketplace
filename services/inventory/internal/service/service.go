package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/phuchoang2603/refurbished-marketplace/services/inventory/internal/database"
	"github.com/phuchoang2603/refurbished-marketplace/shared/err/dberr"

	"github.com/google/uuid"
)

var (
	ErrInvalidProductID  = errors.New("invalid product id")
	ErrInvalidQuantity   = errors.New("invalid quantity")
	ErrInvalidMerchantID = errors.New("invalid merchant id")
	ErrInventoryNotFound = errors.New("inventory not found")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrInvalidBatchSize  = errors.New("invalid product batch size")
)

const (
	ReservationStatusReserved  = "RESERVED"
	ReservationStatusCommitted = "COMMITTED"
	ReservationStatusReleased  = "RELEASED"
	maxStocksByIDs             = 100
)

type Inventory struct {
	ProductID    uuid.UUID
	AvailableQty int32
	ReservedQty  int32
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Service struct {
	db      *sql.DB
	queries *database.Queries
}

func New(db *sql.DB) *Service {
	return &Service{db: db, queries: database.New(db)}
}

func validateProductID(productID uuid.UUID) error {
	if productID == uuid.Nil {
		return ErrInvalidProductID
	}
	return nil
}

func validateNonNegativeQuantity(quantity int32) error {
	if quantity < 0 {
		return ErrInvalidQuantity
	}
	return nil
}

func validatePositiveQuantity(quantity int32) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}
	return nil
}

func mapDBInventory(i database.Inventory) Inventory {
	return Inventory{
		ProductID:    i.ProductID,
		AvailableQty: i.AvailableQty,
		ReservedQty:  i.ReservedQty,
		CreatedAt:    i.CreatedAt,
		UpdatedAt:    i.UpdatedAt,
	}
}

func (s *Service) EnsureStock(ctx context.Context, productID uuid.UUID, initialQty int32) (Inventory, error) {
	if err := validateProductID(productID); err != nil {
		return Inventory{}, err
	}
	if err := validateNonNegativeQuantity(initialQty); err != nil {
		return Inventory{}, err
	}

	created, err := s.queries.CreateInventory(ctx, database.CreateInventoryParams{ProductID: productID, AvailableQty: initialQty})
	if err == nil {
		return mapDBInventory(created), nil
	}

	existing, getErr := s.queries.GetInventoryByProductID(ctx, productID)
	if getErr != nil {
		return Inventory{}, err
	}
	return mapDBInventory(existing), nil
}

func (s *Service) GetInventoryByProductID(ctx context.Context, productID uuid.UUID) (Inventory, error) {
	if err := validateProductID(productID); err != nil {
		return Inventory{}, err
	}
	inv, err := s.queries.GetInventoryByProductID(ctx, productID)
	if err != nil {
		return Inventory{}, dberr.MapErrNoRows(err, ErrInventoryNotFound)
	}
	return mapDBInventory(inv), nil
}

func (s *Service) GetInventoriesByProductIDs(ctx context.Context, ids []uuid.UUID) ([]Inventory, error) {
	if len(ids) > maxStocksByIDs {
		return nil, ErrInvalidBatchSize
	}
	if len(ids) == 0 {
		return []Inventory{}, nil
	}
	seen := make(map[uuid.UUID]struct{}, len(ids))
	unique := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if err := validateProductID(id); err != nil {
			return nil, err
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	rows, err := s.queries.GetInventoriesByProductIDs(ctx, unique)
	if err != nil {
		return nil, err
	}
	out := make([]Inventory, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapDBInventory(row))
	}
	return out, nil
}

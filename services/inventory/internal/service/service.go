package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"refurbished-marketplace/services/inventory/internal/database"

	"github.com/google/uuid"
)

type queryStore interface {
	CreateInventory(ctx context.Context, arg database.CreateInventoryParams) (database.Inventory, error)
	GetInventoryByProductID(ctx context.Context, productID uuid.UUID) (database.Inventory, error)
	ReserveStock(ctx context.Context, arg database.ReserveStockParams) (database.Inventory, error)
	CommitReservation(ctx context.Context, arg database.CommitReservationParams) (database.Inventory, error)
	ReleaseReservation(ctx context.Context, arg database.ReleaseReservationParams) (database.Inventory, error)
}

var (
	ErrInvalidProductID  = errors.New("invalid product id")
	ErrInvalidQuantity   = errors.New("invalid quantity")
	ErrInventoryNotFound = errors.New("inventory not found")
	ErrInsufficientStock = errors.New("insufficient stock")
)

type Inventory struct {
	ProductID    uuid.UUID
	AvailableQty int32
	ReservedQty  int32
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Service struct {
	queries queryStore
}

func New(queries queryStore) *Service {
	return &Service{queries: queries}
}

func (s *Service) CreateInventory(ctx context.Context, productID uuid.UUID, availableQty int32) (Inventory, error) {
	if productID == uuid.Nil {
		return Inventory{}, ErrInvalidProductID
	}
	if availableQty < 0 {
		return Inventory{}, ErrInvalidQuantity
	}

	created, err := s.queries.CreateInventory(ctx, database.CreateInventoryParams{ProductID: productID, AvailableQty: availableQty})
	if err != nil {
		return Inventory{}, err
	}
	return mapDBInventory(created), nil
}

func (s *Service) GetInventoryByProductID(ctx context.Context, productID uuid.UUID) (Inventory, error) {
	if productID == uuid.Nil {
		return Inventory{}, ErrInvalidProductID
	}

	inv, err := s.queries.GetInventoryByProductID(ctx, productID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Inventory{}, ErrInventoryNotFound
		}
		return Inventory{}, err
	}

	return mapDBInventory(inv), nil
}

func (s *Service) ReserveStock(ctx context.Context, productID uuid.UUID, quantity int32) (Inventory, error) {
	if productID == uuid.Nil {
		return Inventory{}, ErrInvalidProductID
	}
	if quantity <= 0 {
		return Inventory{}, ErrInvalidQuantity
	}

	inv, err := s.queries.ReserveStock(ctx, database.ReserveStockParams{ProductID: productID, AvailableQty: quantity})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Inventory{}, ErrInsufficientStock
		}
		return Inventory{}, err
	}

	return mapDBInventory(inv), nil
}

func (s *Service) CommitReservation(ctx context.Context, productID uuid.UUID, quantity int32) (Inventory, error) {
	if productID == uuid.Nil {
		return Inventory{}, ErrInvalidProductID
	}
	if quantity <= 0 {
		return Inventory{}, ErrInvalidQuantity
	}

	inv, err := s.queries.CommitReservation(ctx, database.CommitReservationParams{ProductID: productID, ReservedQty: quantity})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Inventory{}, ErrInventoryNotFound
		}
		return Inventory{}, err
	}

	return mapDBInventory(inv), nil
}

func (s *Service) ReleaseReservation(ctx context.Context, productID uuid.UUID, quantity int32) (Inventory, error) {
	if productID == uuid.Nil {
		return Inventory{}, ErrInvalidProductID
	}
	if quantity <= 0 {
		return Inventory{}, ErrInvalidQuantity
	}

	inv, err := s.queries.ReleaseReservation(ctx, database.ReleaseReservationParams{ProductID: productID, AvailableQty: quantity})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Inventory{}, ErrInventoryNotFound
		}
		return Inventory{}, err
	}

	return mapDBInventory(inv), nil
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

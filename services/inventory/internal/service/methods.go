package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"refurbished-marketplace/services/inventory/internal/database"

	"github.com/google/uuid"
)

type Inventory struct {
	ProductID    uuid.UUID
	AvailableQty int32
	ReservedQty  int32
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (s *Service) CreateInventory(ctx context.Context, productID uuid.UUID, availableQty int32) (Inventory, error) {
	if err := validateProductID(productID); err != nil {
		return Inventory{}, err
	}
	if availableQty < 0 {
		return Inventory{}, ErrInvalidQuantity
	}

	created, err := s.queries.CreateInventory(
		ctx,
		database.CreateInventoryParams{ProductID: productID, AvailableQty: availableQty},
	)
	if err != nil {
		return Inventory{}, err
	}
	return mapDBInventory(created), nil
}

func (s *Service) GetInventoryByProductID(ctx context.Context, productID uuid.UUID) (Inventory, error) {
	if err := validateProductID(productID); err != nil {
		return Inventory{}, err
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

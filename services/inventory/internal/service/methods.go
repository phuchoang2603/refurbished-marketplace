package service

import (
	"context"
	"time"

	"refurbished-marketplace/services/inventory/internal/database"
	"refurbished-marketplace/shared/dberrors"

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
		if dberrors.IsNoRows(err) {
			return Inventory{}, ErrInventoryNotFound
		}
		return Inventory{}, err
	}

	return mapDBInventory(inv), nil
}

func (s *Service) ReserveStock(ctx context.Context, productID uuid.UUID, quantity int32) (Inventory, error) {
	if err := validateProductID(productID); err != nil {
		return Inventory{}, err
	}
	if err := validatePositiveQuantity(quantity); err != nil {
		return Inventory{}, err
	}

	inv, err := s.queries.ReserveStock(ctx, database.ReserveStockParams{ProductID: productID, Quantity: quantity})
	if err != nil {
		if dberrors.IsNoRows(err) {
			if _, getErr := s.queries.GetInventoryByProductID(ctx, productID); dberrors.IsNoRows(getErr) {
				return Inventory{}, ErrInventoryNotFound
			} else if getErr != nil {
				return Inventory{}, getErr
			}
			return Inventory{}, ErrInsufficientStock
		}
		return Inventory{}, err
	}

	return mapDBInventory(inv), nil
}

func (s *Service) CommitReservation(ctx context.Context, productID uuid.UUID, quantity int32) (Inventory, error) {
	if err := validateProductID(productID); err != nil {
		return Inventory{}, err
	}
	if err := validatePositiveQuantity(quantity); err != nil {
		return Inventory{}, err
	}

	inv, err := s.queries.CommitReservation(ctx, database.CommitReservationParams{ProductID: productID, Quantity: quantity})
	if err != nil {
		if dberrors.IsNoRows(err) {
			return Inventory{}, ErrInventoryNotFound
		}
		return Inventory{}, err
	}

	return mapDBInventory(inv), nil
}

func (s *Service) ReleaseReservation(ctx context.Context, productID uuid.UUID, quantity int32) (Inventory, error) {
	if err := validateProductID(productID); err != nil {
		return Inventory{}, err
	}
	if err := validatePositiveQuantity(quantity); err != nil {
		return Inventory{}, err
	}

	inv, err := s.queries.ReleaseReservation(ctx, database.ReleaseReservationParams{ProductID: productID, Quantity: quantity})
	if err != nil {
		if dberrors.IsNoRows(err) {
			return Inventory{}, ErrInventoryNotFound
		}
		return Inventory{}, err
	}

	return mapDBInventory(inv), nil
}

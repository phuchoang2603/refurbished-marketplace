package service

import (
	"context"
	"time"

	"refurbished-marketplace/services/products/internal/database"
	shareddb "refurbished-marketplace/shared/db"

	"github.com/google/uuid"
)

type Product struct {
	ID           uuid.UUID
	Name         string
	Description  string
	PriceCents   int64
	MerchantID   uuid.UUID
	AvailableQty *int32
	ReservedQty  *int32
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Inventory struct {
	ProductID    uuid.UUID
	AvailableQty int32
	ReservedQty  int32
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (s *Service) CreateProduct(ctx context.Context, name, description string, priceCents int64, merchantID uuid.UUID, initialStock int32) (Product, error) {
	cleanName := normalizeProductName(name)
	if cleanName == "" {
		return Product{}, ErrInvalidProductName
	}

	desc := normalizeProductDescription(description, cleanName)

	if priceCents <= 0 {
		return Product{}, ErrInvalidPrice
	}
	if merchantID == uuid.Nil {
		return Product{}, ErrInvalidMerchantID
	}
	if err := validateNonNegativeQuantity(initialStock); err != nil {
		return Product{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Product{}, err
	}
	q := s.queries.WithTx(tx)
	defer func() {
		_ = tx.Rollback()
	}()

	created, err := q.CreateProduct(ctx, database.CreateProductParams{
		ID:          uuid.New(),
		Name:        cleanName,
		Description: desc,
		PriceCents:  priceCents,
		MerchantID:  merchantID,
	})
	if err != nil {
		return Product{}, err
	}

	if _, err := q.CreateInventory(ctx, database.CreateInventoryParams{
		ProductID:    created.ID,
		AvailableQty: initialStock,
	}); err != nil {
		return Product{}, err
	}

	if err := tx.Commit(); err != nil {
		return Product{}, err
	}

	return mapDBProduct(created), nil
}

func (s *Service) GetProductByID(ctx context.Context, id uuid.UUID) (Product, error) {
	if err := validateProductID(id); err != nil {
		return Product{}, err
	}

	p, err := s.queries.GetProductByID(ctx, id)
	if err != nil {
		return Product{}, mapProductNotFound(err)
	}

	return mapDBProductRow(p), nil
}

func (s *Service) ListProducts(ctx context.Context, limit, offset int32) ([]Product, error) {
	if err := validateListPagination(limit, offset); err != nil {
		return nil, err
	}

	rows, err := s.queries.ListProducts(ctx, database.ListProductsParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}

	result := make([]Product, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapDBProduct(row))
	}
	return result, nil
}

func (s *Service) GetInventoryByProductID(ctx context.Context, productID uuid.UUID) (Inventory, error) {
	if err := validateProductID(productID); err != nil {
		return Inventory{}, err
	}

	inv, err := s.queries.GetInventoryByProductID(ctx, productID)
	if err != nil {
		return Inventory{}, shareddb.MapErrNoRows(err, ErrInventoryNotFound)
	}

	return mapDBInventory(inv), nil
}

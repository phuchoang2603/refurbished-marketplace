package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"refurbished-marketplace/services/products/internal/database"

	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID
	OwnerUserID uuid.UUID
	Name        string
	Description string
	PriceCents  int64
	Stock       int32
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (s *Service) CreateProduct(ctx context.Context, ownerUserID uuid.UUID, name, description string, priceCents int64, stock int32) (Product, error) {
	if ownerUserID == uuid.Nil {
		return Product{}, ErrInvalidOwner
	}

	cleanName := strings.TrimSpace(name)
	if cleanName == "" {
		return Product{}, ErrInvalidProductName
	}

	desc := strings.TrimSpace(description)
	if desc == "" {
		desc = cleanName
	}

	if priceCents <= 0 {
		return Product{}, ErrInvalidPrice
	}

	if stock < 0 {
		return Product{}, ErrInvalidStock
	}

	created, err := s.queries.CreateProduct(ctx, database.CreateProductParams{
		ID:          uuid.New(),
		OwnerUserID: ownerUserID,
		Name:        cleanName,
		Description: desc,
		PriceCents:  priceCents,
		Stock:       stock,
	})
	if err != nil {
		return Product{}, err
	}

	return mapDBProduct(created), nil
}

func (s *Service) GetProductByID(ctx context.Context, id uuid.UUID) (Product, error) {
	p, err := s.queries.GetProductByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Product{}, ErrProductNotFound
		}
		return Product{}, err
	}

	return mapDBProduct(p), nil
}

func (s *Service) ListProducts(ctx context.Context, limit, offset int32) ([]Product, error) {
	if limit <= 0 || limit > 100 {
		return nil, ErrInvalidListLimit
	}
	if offset < 0 {
		return nil, ErrInvalidListOffset
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

func (s *Service) UpdateProduct(ctx context.Context, id uuid.UUID, ownerUserID uuid.UUID, in UpdateProductInput) (Product, error) {
	if id == uuid.Nil {
		return Product{}, ErrProductNotFound
	}
	if ownerUserID == uuid.Nil {
		return Product{}, ErrInvalidOwner
	}

	name := normalizeOptionalText(in.Name)
	description := normalizeOptionalText(in.Description)

	if name != nil && *name == "" {
		return Product{}, ErrInvalidProductName
	}
	if in.PriceCents != nil && *in.PriceCents <= 0 {
		return Product{}, ErrInvalidPrice
	}
	if in.Stock != nil && *in.Stock < 0 {
		return Product{}, ErrInvalidStock
	}

	updated, err := s.queries.UpdateProductByIDAndOwner(ctx, database.UpdateProductByIDAndOwnerParams{
		ID:          id,
		OwnerUserID: ownerUserID,
		Name:        toNullString(name),
		Description: toNullString(description),
		PriceCents:  toNullInt64(in.PriceCents),
		Stock:       toNullInt32(in.Stock),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Product{}, ErrForbiddenProduct
		}
		return Product{}, err
	}

	return mapDBProduct(updated), nil
}

func (s *Service) DeleteProduct(ctx context.Context, id uuid.UUID, ownerUserID uuid.UUID) error {
	if id == uuid.Nil {
		return ErrProductNotFound
	}
	if ownerUserID == uuid.Nil {
		return ErrInvalidOwner
	}

	rows, err := s.queries.DeleteProductByIDAndOwner(ctx, database.DeleteProductByIDAndOwnerParams{ID: id, OwnerUserID: ownerUserID})
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrForbiddenProduct
	}

	return nil
}

func normalizeOptionalText(v *string) *string {
	if v == nil {
		return nil
	}
	vv := strings.TrimSpace(*v)
	return &vv
}

func toNullString(v *string) sql.NullString {
	if v == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *v, Valid: true}
}

func toNullInt64(v *int64) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *v, Valid: true}
}

func toNullInt32(v *int32) sql.NullInt32 {
	if v == nil {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: *v, Valid: true}
}

func mapDBProduct(p database.Product) Product {
	return Product{
		ID:          p.ID,
		OwnerUserID: p.OwnerUserID,
		Name:        p.Name,
		Description: p.Description,
		PriceCents:  p.PriceCents,
		Stock:       p.Stock,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

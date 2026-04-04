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
	Name        string
	Description string
	PriceCents  int64
	TerminalID  uuid.UUID
	XPos        float64
	YPos        float64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (s *Service) CreateProduct(ctx context.Context, name, description string, priceCents int64, terminalID uuid.UUID, xPos, yPos float64) (Product, error) {
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
	if terminalID == uuid.Nil {
		return Product{}, ErrInvalidTerminalID
	}

	created, err := s.queries.CreateProduct(ctx, database.CreateProductParams{
		ID:          uuid.New(),
		Name:        cleanName,
		Description: desc,
		PriceCents:  priceCents,
		TerminalID:  terminalID,
		XPos:        xPos,
		YPos:        yPos,
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

func mapDBProduct(p database.Product) Product {
	return Product{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		PriceCents:  p.PriceCents,
		TerminalID:  p.TerminalID,
		XPos:        p.XPos,
		YPos:        p.YPos,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

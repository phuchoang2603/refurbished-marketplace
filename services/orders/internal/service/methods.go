package service

import (
	"context"
	"time"

	sharedlog "refurbished-marketplace/shared/observe/log"

	"refurbished-marketplace/services/orders/internal/database"
	"refurbished-marketplace/shared/err/dberr"

	"github.com/google/uuid"
)

type OrderItemInput struct {
	ProductID      uuid.UUID
	Quantity       int32
	UnitPriceCents int64
}

type Order struct {
	ID          uuid.UUID
	BuyerUserID uuid.UUID
	MerchantID  uuid.UUID
	Status      string
	TotalCents  int64
	Items       []OrderItem
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type OrderItem struct {
	ID             uuid.UUID
	OrderID        uuid.UUID
	ProductID      uuid.UUID
	Quantity       int32
	UnitPriceCents int64
	LineTotalCents int64
	CreatedAt      time.Time
}

func (s *Service) CreateOrder(ctx context.Context, buyerUserID, merchantID uuid.UUID, items []OrderItemInput, totalCents int64, checkoutIntentKey string) (Order, error) {
	checkoutIntentKey = normalizeCheckoutIntentKey(checkoutIntentKey)
	if err := validateCreateOrderInput(buyerUserID, merchantID, items, totalCents, checkoutIntentKey); err != nil {
		return Order{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Order{}, err
	}
	q := s.queries.WithTx(tx)
	defer func() {
		_ = tx.Rollback()
	}()

	created, err := q.CreateOrder(ctx, database.CreateOrderParams{
		ID:                           uuid.New(),
		BuyerUserID:                  buyerUserID,
		CheckoutIntentIdempotencyKey: dberr.OptionalNullString(checkoutIntentKey),
		MerchantID:                   merchantID,
		Status:                       OrderStatusPending,
		TotalCents:                   totalCents,
	})
	if err != nil {
		if dberr.IsUniqueViolation(err) {
			_ = tx.Rollback()
			existing, getErr := s.loadOrderByCheckoutIntent(ctx, s.queries, buyerUserID, checkoutIntentKey)
			if getErr != nil {
				return Order{}, getErr
			}
			if !orderMatchesCreateInput(existing, merchantID, items, totalCents) {
				return Order{}, ErrCheckoutIntentConflict
			}
			return existing, nil
		}
		return Order{}, err
	}

	orderItems, err := createOrderItems(ctx, q, created.ID, items)
	if err != nil {
		return Order{}, err
	}

	createdOrder := mapOrderFields(created.ID, created.BuyerUserID, created.MerchantID, created.Status, created.TotalCents, created.CreatedAt, created.UpdatedAt)
	createdOrder.Items = orderItems
	if err := createOrderOutbox(ctx, q, createdOrder); err != nil {
		return Order{}, err
	}

	if err := tx.Commit(); err != nil {
		return Order{}, err
	}

	sharedlog.InfoContext(
		ctx, "order created",
		sharedlog.KeyOrderID, createdOrder.ID.String(),
		sharedlog.KeyBuyerUserID, createdOrder.BuyerUserID.String(),
		sharedlog.KeyMerchantID, createdOrder.MerchantID.String(),
		"total_cents", createdOrder.TotalCents,
		"item_count", len(createdOrder.Items),
	)
	return createdOrder, nil
}

func (s *Service) loadOrderByCheckoutIntent(ctx context.Context, q *database.Queries, buyerUserID uuid.UUID, checkoutIntentKey string) (Order, error) {
	row, err := q.GetOrderByBuyerAndCheckoutIntentKey(ctx, database.GetOrderByBuyerAndCheckoutIntentKeyParams{
		BuyerUserID:                  buyerUserID,
		CheckoutIntentIdempotencyKey: dberr.OptionalNullString(checkoutIntentKey),
	})
	if err != nil {
		return Order{}, dberr.MapErrNoRows(err, ErrOrderNotFound)
	}
	orders, err := loadOrdersWithItems(ctx, q, []Order{mapOrderFields(row.ID, row.BuyerUserID, row.MerchantID, row.Status, row.TotalCents, row.CreatedAt, row.UpdatedAt)})
	if err != nil {
		return Order{}, err
	}
	if len(orders) == 0 {
		return Order{}, ErrOrderNotFound
	}
	return orders[0], nil
}

func orderMatchesCreateInput(existing Order, merchantID uuid.UUID, items []OrderItemInput, totalCents int64) bool {
	if existing.MerchantID != merchantID || existing.TotalCents != totalCents || len(existing.Items) != len(items) {
		return false
	}
	for i, item := range items {
		if existing.Items[i].ProductID != item.ProductID {
			return false
		}
		if existing.Items[i].Quantity != item.Quantity {
			return false
		}
		if existing.Items[i].UnitPriceCents != item.UnitPriceCents {
			return false
		}
	}
	return true
}

func (s *Service) GetOrderByID(ctx context.Context, id uuid.UUID) (Order, error) {
	if id == uuid.Nil {
		return Order{}, ErrOrderNotFound
	}

	got, err := s.queries.GetOrderByID(ctx, id)
	if err != nil {
		return Order{}, dberr.MapErrNoRows(err, ErrOrderNotFound)
	}

	orders, err := loadOrdersWithItems(ctx, s.queries, []Order{mapOrderFields(got.ID, got.BuyerUserID, got.MerchantID, got.Status, got.TotalCents, got.CreatedAt, got.UpdatedAt)})
	if err != nil {
		return Order{}, err
	}
	if len(orders) == 0 {
		return Order{}, ErrOrderNotFound
	}
	return orders[0], nil
}

func (s *Service) ListOrdersByBuyer(ctx context.Context, buyerUserID uuid.UUID, limit, offset int32) ([]Order, error) {
	if buyerUserID == uuid.Nil {
		return nil, ErrInvalidBuyerID
	}
	if err := validateListPagination(limit, offset); err != nil {
		return nil, err
	}

	rows, err := s.queries.ListOrdersByBuyer(ctx, database.ListOrdersByBuyerParams{BuyerUserID: buyerUserID, Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}

	orders := make([]Order, 0, len(rows))
	for _, row := range rows {
		orders = append(orders, mapOrderFields(row.ID, row.BuyerUserID, row.MerchantID, row.Status, row.TotalCents, row.CreatedAt, row.UpdatedAt))
	}
	return loadOrdersWithItems(ctx, s.queries, orders)
}

func (s *Service) UpdateOrderStatus(ctx context.Context, id uuid.UUID, status string) (Order, error) {
	if err := s.updateOrderStatusOnly(ctx, id, status); err != nil {
		return Order{}, err
	}

	got, err := s.queries.GetOrderByID(ctx, id)
	if err != nil {
		return Order{}, dberr.MapErrNoRows(err, ErrOrderNotFound)
	}

	orders, err := loadOrdersWithItems(ctx, s.queries, []Order{mapOrderFields(got.ID, got.BuyerUserID, got.MerchantID, got.Status, got.TotalCents, got.CreatedAt, got.UpdatedAt)})
	if err != nil {
		return Order{}, err
	}
	if len(orders) == 0 {
		return Order{}, ErrOrderNotFound
	}
	return orders[0], nil
}

func (s *Service) updateOrderStatusOnly(ctx context.Context, id uuid.UUID, status string) error {
	return s.updateOrderStatusWithQueries(ctx, s.queries, id, status)
}

func (s *Service) updateOrderStatusWithQueries(ctx context.Context, q *database.Queries, id uuid.UUID, status string) error {
	if id == uuid.Nil {
		return ErrOrderNotFound
	}
	normalizedStatus, err := validateOrderStatus(status)
	if err != nil {
		return ErrInvalidStatus
	}

	_, err = q.UpdateOrderStatus(ctx, database.UpdateOrderStatusParams{ID: id, Status: normalizedStatus})
	if err != nil {
		return dberr.MapErrNoRows(err, ErrOrderNotFound)
	}
	return nil
}

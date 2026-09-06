package service

import (
	"context"
	"time"

	sharedlog "github.com/phuchoang2603/refurbished-marketplace/shared/observe/log"

	"github.com/phuchoang2603/refurbished-marketplace/services/orders/internal/database"
	"github.com/phuchoang2603/refurbished-marketplace/shared/err/dberr"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func (s *Service) CreateOrder(ctx context.Context, buyerUserID, merchantID uuid.UUID, items []OrderItemInput, totalCents int64, idempotencyKey uuid.UUID) (Order, error) {
	if err := validateCreateOrderInput(buyerUserID, merchantID, idempotencyKey, items, totalCents); err != nil {
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
		ID:             uuid.New(),
		BuyerUserID:    buyerUserID,
		MerchantID:     merchantID,
		Status:         OrderStatusPending,
		TotalCents:     totalCents,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		_ = tx.Rollback()
		if dberr.IsUniqueViolation(err) {
			return s.replayCreateOrder(ctx, buyerUserID, merchantID, items, totalCents, idempotencyKey)
		}
		return Order{}, err
	}

	orderItems, err := createOrderItems(ctx, q, created.ID, items)
	if err != nil {
		return Order{}, err
	}

	createdOrder := mapDBOrder(created)
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
	if err := s.reserveStock(ctx, createdOrder); err != nil {
		return Order{}, err
	}
	return createdOrder, nil
}

func (s *Service) replayCreateOrder(ctx context.Context, buyerUserID, merchantID uuid.UUID, items []OrderItemInput, totalCents int64, idempotencyKey uuid.UUID) (Order, error) {
	existing, err := s.queries.GetOrderByBuyerIdempotencyKey(ctx, database.GetOrderByBuyerIdempotencyKeyParams{
		BuyerUserID:    buyerUserID,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return Order{}, dberr.MapErrNoRows(err, ErrOrderNotFound)
	}
	orders, err := loadOrdersWithItems(ctx, s.queries, []Order{mapDBOrder(existing)})
	if err != nil {
		return Order{}, err
	}
	if len(orders) == 0 {
		return Order{}, ErrOrderNotFound
	}
	order := orders[0]
	if order.MerchantID != merchantID || order.TotalCents != totalCents || !orderItemsMatch(order.Items, items) {
		return Order{}, ErrIdempotencyConflict
	}
	if order.Status == OrderStatusFailed {
		return Order{}, ErrOrderNotPayable
	}
	if order.Status == OrderStatusPending {
		if err := s.reserveStock(ctx, order); err != nil {
			return Order{}, err
		}
	}
	return order, nil
}

func orderItemsMatch(existing []OrderItem, input []OrderItemInput) bool {
	if len(existing) != len(input) {
		return false
	}
	type line struct {
		qty   int32
		price int64
	}
	want := make(map[uuid.UUID]line, len(input))
	for _, item := range input {
		want[item.ProductID] = line{qty: item.Quantity, price: item.UnitPriceCents}
	}
	for _, item := range existing {
		got, ok := want[item.ProductID]
		if !ok || got.qty != item.Quantity || got.price != item.UnitPriceCents {
			return false
		}
	}
	return true
}

func (s *Service) reserveStock(ctx context.Context, order Order) error {
	if s.stock == nil {
		return nil
	}
	err := s.stock.ReserveStock(ctx, order.ID, order.MerchantID, order.TotalCents, orderItemsToInput(order.Items))
	if err == nil {
		return nil
	}
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.FailedPrecondition:
			return ErrInsufficientStock
		case codes.InvalidArgument:
			return ErrInvalidQuantity
		}
	}
	return err
}

func orderItemsToInput(items []OrderItem) []OrderItemInput {
	out := make([]OrderItemInput, 0, len(items))
	for _, item := range items {
		out = append(out, OrderItemInput{
			ProductID:      item.ProductID,
			Quantity:       item.Quantity,
			UnitPriceCents: item.UnitPriceCents,
		})
	}
	return out
}

func (s *Service) GetOrderByID(ctx context.Context, id uuid.UUID) (Order, error) {
	if id == uuid.Nil {
		return Order{}, ErrOrderNotFound
	}

	got, err := s.queries.GetOrderByID(ctx, id)
	if err != nil {
		return Order{}, dberr.MapErrNoRows(err, ErrOrderNotFound)
	}

	orders, err := loadOrdersWithItems(ctx, s.queries, []Order{mapDBOrder(got)})
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
		orders = append(orders, mapDBOrder(row))
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

	orders, err := loadOrdersWithItems(ctx, s.queries, []Order{mapDBOrder(got)})
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

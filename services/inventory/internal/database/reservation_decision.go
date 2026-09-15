package database

import (
	"context"

	"github.com/google/uuid"
)

const lockInventoryReservationOrder = `SELECT pg_advisory_xact_lock(hashtext($1::text), 1)`

func (q *Queries) LockInventoryReservationOrder(ctx context.Context, orderID uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, lockInventoryReservationOrder, orderID.String())
	return err
}

const inventoryInboxExists = `SELECT EXISTS (SELECT 1 FROM inventory_inbox WHERE message_id = $1)`

func (q *Queries) InventoryInboxExists(ctx context.Context, messageID string) (bool, error) {
	row := q.db.QueryRowContext(ctx, inventoryInboxExists, messageID)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}

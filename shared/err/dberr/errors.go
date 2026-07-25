package db

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

func MapErrNoRows(err, notFound error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return notFound
	}
	return err
}

func IsUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}

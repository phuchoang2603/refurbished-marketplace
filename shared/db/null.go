package db

import (
	"database/sql"
	"time"
)

func OptionalNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func OptionalNullTime(t time.Time) sql.NullTime {
	return sql.NullTime{Time: t, Valid: !t.IsZero()}
}

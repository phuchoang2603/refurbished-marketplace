package messaging

import "time"

type AccountCreated struct {
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

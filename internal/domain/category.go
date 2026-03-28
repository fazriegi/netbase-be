package domain

import "github.com/google/uuid"

type Category struct {
	UserID   uuid.UUID `db:"user_id"`
	Name     string    `db:"name"`
	BaseType string    `db:"base_type"`
}

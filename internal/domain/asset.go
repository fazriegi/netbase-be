package domain

import (
	"github.com/fazriegi/fintrack-be/pkg"
	"github.com/google/uuid"
)

type Asset struct {
	ID           uuid.UUID `db:"id" json:"id"`
	UserId       uuid.UUID `db:"user_id" json:"user_id"`
	Category     string    `db:"category" json:"category"`
	Name         string    `db:"name" json:"name"`
	CurrentValue string    `db:"current_value" json:"current_value"`
	Detail       any       `db:"details" json:"detail"`
	IsActive     bool      `db:"is_active" json:"is_active"`
}

type ListAssetRequest struct {
	pkg.PaginationRequest
	UserId uuid.UUID
	Search string `query:"search"`
}

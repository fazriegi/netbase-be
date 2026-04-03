package domain

import (
	"github.com/fazriegi/fintrack-be/pkg"
	"github.com/google/uuid"
)

type Asset struct {
	ID           uuid.UUID `db:"id" json:"id"`
	UserId       uuid.UUID `db:"user_id" json:"user_id"`
	CategoryID   uuid.UUID `db:"category_id" json:"category_id"`
	Category     string    `db:"category" json:"category"`
	CategoryType string    `db:"category_type" json:"category_type"`
	Name         string    `db:"name" json:"name"`
	CurrentValue string    `db:"current_value" json:"current_value"`
	Detail       any       `db:"details" json:"detail"`
	IsActive     bool      `db:"is_active" json:"is_active"`
}

type ListAssetRequest struct {
	pkg.PaginationRequest
	UserId   uuid.UUID
	Name     string `query:"name"`
	Category string `query:"category"`
	IsActive *bool  `query:"is_active"`
}

type ListAssetCategoryResponse struct {
	Categories *[]Category `json:"categories"`
}

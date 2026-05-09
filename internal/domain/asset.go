package domain

import (
	"time"

	"github.com/fazriegi/fintrack-be/pkg"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Asset struct {
	ID           uuid.UUID       `db:"id" json:"id"`
	UserId       uuid.UUID       `db:"user_id" json:"user_id"`
	CategoryID   uuid.UUID       `db:"category_id" json:"category_id"`
	Category     string          `db:"category" json:"category"`
	CategoryType string          `db:"category_type" json:"category_type"`
	Name         string          `db:"name" json:"name"`
	CurrentValue decimal.Decimal `db:"current_value" json:"current_value"`
	Details      any             `db:"details" json:"details"`
	IsActive     bool            `db:"is_active" json:"is_active"`
	CreatedAt    time.Time       `db:"created_at" json:"-"`
}

type AssetDB struct {
	ID           uuid.UUID       `db:"id"`
	UserId       uuid.UUID       `db:"user_id"`
	CategoryID   uuid.UUID       `db:"category_id"`
	Name         string          `db:"name"`
	CurrentValue decimal.Decimal `db:"current_value"`
	Details      any             `db:"details"`
	IsActive     bool            `db:"is_active"`
	CreatedAt    time.Time       `db:"created_at"`
	UpdatedAt    time.Time       `db:"updated_at"`
}

type ListAssetRequest struct {
	pkg.PaginationRequest
	UserId   uuid.UUID
	Name     string `query:"name"`
	Category string `query:"category"`
	IsActive *bool  `query:"is_active"`
}

type ListAssetResponse struct {
	ID           uuid.UUID       `json:"id"`
	Category     string          `json:"category"`
	Name         string          `json:"name"`
	CurrentValue decimal.Decimal `json:"current_value"`
	IsActive     bool            `json:"is_active"`
}

type ListAssetCategoryResponse struct {
	Categories *[]Category `json:"categories"`
}

type GetAssetByIDResponse struct {
	ID           uuid.UUID       `json:"id"`
	CategoryID   uuid.UUID       `json:"category_id"`
	Category     string          `json:"category"`
	CategoryType string          `json:"category_type"`
	Name         string          `json:"name"`
	CurrentValue decimal.Decimal `json:"current_value"`
	Details      any             `json:"details"`
	IsActive     bool            `json:"is_active"`
}

type CreateAsset struct {
	UserId       uuid.UUID
	Name         string           `json:"name" validate:"required"`
	CategoryID   uuid.UUID        `json:"category_id" validate:"required"`
	CurrentValue *decimal.Decimal `json:"current_value" validate:"required"`
	Details      any              `json:"details" validate:"required"`
	IsActive     *bool            `json:"is_active" validate:"required"`
	CategoryType string           `json:"category_type" validate:"required"`
}

type UpdateAsset struct {
	UserId       uuid.UUID
	ID           uuid.UUID
	Name         string           `json:"name" validate:"required"`
	CategoryID   uuid.UUID        `json:"category_id" validate:"required"`
	CurrentValue *decimal.Decimal `json:"current_value" validate:"required"`
	Details      any              `json:"details" validate:"required"`
	IsActive     *bool            `json:"is_active" validate:"required"`
	CategoryType string           `json:"category_type" validate:"required"`
}

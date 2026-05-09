package domain

import (
	"time"

	"github.com/fazriegi/fintrack-be/pkg"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type LiabilityDB struct {
	ID               uuid.UUID       `db:"id"`
	UserId           uuid.UUID       `db:"user_id"`
	CategoryID       uuid.UUID       `db:"category_id"`
	Name             string          `db:"name"`
	PrincipalAmount  decimal.Decimal `db:"principal_amount"`
	RemainingBalance decimal.Decimal `db:"remaining_balance"`
	Details          any             `db:"details"`
	CreatedAt        time.Time       `db:"created_at"`
	UpdatedAt        time.Time       `db:"updated_at"`
}

type Liability struct {
	ID               uuid.UUID       `db:"id" json:"id"`
	UserId           uuid.UUID       `db:"user_id" json:"user_id"`
	CategoryID       uuid.UUID       `db:"category_id" json:"category_id"`
	Category         string          `db:"category" json:"category"`
	Name             string          `db:"name" json:"name"`
	CategoryType     string          `db:"category_type" json:"category_type"`
	RemainingBalance decimal.Decimal `db:"remaining_balance" json:"remaining_balance"`
	PrincipalAmount  decimal.Decimal `db:"principal_amount" json:"principal_amount"`
	Details          any             `db:"details" json:"details"`
	CreatedAt        time.Time       `db:"created_at" json:"-"`
}

type CreateLiability struct {
	ID               uuid.UUID
	UserId           uuid.UUID
	Name             string           `json:"name" validate:"required"`
	CategoryID       uuid.UUID        `json:"category_id" validate:"required"`
	PrincipalAmount  *decimal.Decimal `json:"principal_amount" validate:"required"`
	RemainingBalance *decimal.Decimal `json:"remaining_balance" validate:"required"`
	Details          any              `json:"details" validate:"required"`
	CategoryType     string           `json:"category_type" validate:"required"`
}

type ListLiabilityRequest struct {
	pkg.PaginationRequest
	UserId   uuid.UUID
	Name     string `query:"name"`
	Category string `query:"category"`
}

type ListLiabilityResponse struct {
	ID               uuid.UUID       `json:"id"`
	Category         string          `json:"category"`
	Name             string          `json:"name"`
	RemainingBalance decimal.Decimal `json:"remaining_balance"`
}

type GetLiabilityByIDResponse struct {
	ID               uuid.UUID       `json:"id"`
	CategoryID       uuid.UUID       `json:"category_id"`
	Category         string          `json:"category"`
	Name             string          `json:"name"`
	CategoryType     string          `json:"category_type"`
	PrincipalAmount  decimal.Decimal `json:"principal_amount"`
	RemainingBalance decimal.Decimal `json:"remaining_balance"`
	Details          any             `json:"details"`
}

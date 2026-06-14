package domain

import (
	"context"
	"time"

	"github.com/fazriegi/fintrack-be/pkg"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionDB struct {
	ID              uuid.UUID       `db:"id"`
	UserID          uuid.UUID       `db:"user_id"`
	AssetID         *uuid.UUID      `db:"asset_id"`
	LiabilityID     *uuid.UUID      `db:"liability_id"`
	CategoryID      uuid.UUID       `db:"category_id"`
	Amount          decimal.Decimal `db:"amount"`
	TransactionDate time.Time       `db:"transaction_date"`
	Notes           *string         `db:"notes"`
	CreatedAt       time.Time       `db:"created_at"`
	UpdatedAt       time.Time       `db:"updated_at"`
}

type Transaction struct {
	ID              uuid.UUID       `db:"id" json:"id"`
	UserID          uuid.UUID       `db:"user_id" json:"user_id"`
	AssetID         *uuid.UUID      `db:"asset_id" json:"asset_id"`
	AssetName       *string         `db:"asset_name" json:"asset_name"`
	LiabilityID     *uuid.UUID      `db:"liability_id" json:"liability_id"`
	LiabilityName   *string         `db:"liability_name" json:"liability_name"`
	CategoryID      uuid.UUID       `db:"category_id" json:"category_id"`
	CategoryName    string          `db:"category_name" json:"category_name"`
	CategoryType    string          `db:"category_type" json:"category_type"`
	Amount          decimal.Decimal `db:"amount" json:"amount"`
	TransactionDate time.Time       `db:"transaction_date" json:"transaction_date"`
	Notes           *string         `db:"notes" json:"notes"`
	CreatedAt       time.Time       `db:"created_at" json:"-"`
}

type CreateTransaction struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	AssetID         *uuid.UUID       `json:"asset_id"`
	LiabilityID     *uuid.UUID       `json:"liability_id"`
	CategoryID      uuid.UUID        `json:"category_id" validate:"required"`
	Amount          *decimal.Decimal `json:"amount" validate:"required"`
	TransactionDate string           `json:"transaction_date" validate:"required"`
	Notes           *string          `json:"notes"`
}

type ListTransactionRequest struct {
	pkg.PaginationRequest
	UserID       uuid.UUID
	CategoryName string `query:"category_name"`
	Notes        string `query:"notes"`
	FilterType   string `query:"filter_type"` // "week", "month", "year", "range"
	DateStr      string `query:"date"`        // reference date YYYY-MM-DD
	StartDateStr string `query:"start_date"`  // range start date YYYY-MM-DD
	EndDateStr   string `query:"end_date"`    // range end date YYYY-MM-DD
}

type ListTransactionResponse struct {
	ID              uuid.UUID       `json:"id"`
	AssetID         *uuid.UUID      `json:"asset_id"`
	AssetName       *string         `json:"asset_name"`
	LiabilityID     *uuid.UUID      `json:"liability_id"`
	LiabilityName   *string         `json:"liability_name"`
	CategoryID      uuid.UUID       `json:"category_id"`
	CategoryName    string          `json:"category_name"`
	CategoryType    string          `json:"category_type"`
	Amount          decimal.Decimal `json:"amount"`
	TransactionDate time.Time       `json:"transaction_date"`
	Notes           *string         `json:"notes"`
}

type ListCategoryRequest struct {
	UserID   uuid.UUID
	BaseType string `query:"base_type"` // "income", "expense"
	Search   string `query:"search"`
}

type TransactionRepository interface {
	ListCategory(ctx context.Context, req *ListCategoryRequest) (*[]Category, error)
	GetCategoryByID(ctx context.Context, id, userID uuid.UUID) (*Category, error)
	InsertCategory(ctx context.Context, category *Category) error
	DeleteCategory(ctx context.Context, id, userID uuid.UUID) error
	List(ctx context.Context, req *ListTransactionRequest) (*[]Transaction, int, error)
	GetByID(ctx context.Context, id, userID uuid.UUID) (*Transaction, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
	Insert(ctx context.Context, data *TransactionDB) error
	Update(ctx context.Context, data *TransactionDB) error
}

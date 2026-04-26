package domain

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Category struct {
	Id       uuid.UUID `db:"id" json:"id"`
	UserID   uuid.UUID `db:"user_id" json:"-"`
	Name     string    `db:"name" json:"name"`
	BaseType string    `db:"base_type" json:"base_type"`
}

type LiquidAsset struct {
	PlatformName   string          `json:"platform_name"`
	AccountName    string          `json:"account_name"`
	AccountNumber  string          `json:"account_number"`
	InterestRatePA decimal.Decimal `json:"interest_rate_pa"`
}

type InvestmentAsset struct {
	PlatformName string          `json:"platform_name"`
	TickerSymbol string          `json:"ticker_symbol"`
	AveragePrice decimal.Decimal `json:"average_price"`
	Quantity     decimal.Decimal `json:"quantity"`
}

type PhysicalAsset struct {
	Model         string          `json:"model"`
	PurchaseYear  int             `json:"purchase_year"`
	PurchasePrice decimal.Decimal `json:"purchase_price"`
}

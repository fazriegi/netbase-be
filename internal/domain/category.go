package domain

import (
	"time"

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
	PlatformName   string          `json:"platform_name" validate:"required"`
	AccountName    string          `json:"account_name"`
	AccountNumber  string          `json:"account_number"`
	InterestRatePA decimal.Decimal `json:"interest_rate_pa"`
}

type InvestmentAsset struct {
	PlatformName string          `json:"platform_name" validate:"required"`
	TickerSymbol string          `json:"ticker_symbol" validate:"required"`
	AveragePrice decimal.Decimal `json:"average_price" validate:"required"`
	Quantity     decimal.Decimal `json:"quantity" validate:"required"`
}

type PhysicalAsset struct {
	Model         string          `json:"model" validate:"required"`
	PurchaseYear  int             `json:"purchase_year" validate:"required"`
	PurchasePrice decimal.Decimal `json:"purchase_price"`
}

type ShortTermLiability struct {
	CreditLimit   decimal.Decimal `json:"credit_limit"`
	StatementDate int             `json:"statement_date" validate:"required,number,lte=31,gte=1"`
	DueDate       int             `json:"due_date" validate:"required,number,lte=31,gte=1"`
	InterestRate  decimal.Decimal `json:"interest_rate"`
}

type LongTermLiability struct {
	MonthlyInstallment decimal.Decimal `json:"monthly_installment" validate:"required"`
	Tenor              int             `json:"tenor" validate:"required"`
	DueDate            int             `json:"due_date" validate:"required,number,lte=31,gte=1"`
	InterestRatePA     decimal.Decimal `json:"interest_rate_pa" `
	StartDate          time.Time       `json:"start_date" validate:"required"`
}

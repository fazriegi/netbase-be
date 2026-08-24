package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Networth struct {
	ID               uuid.UUID       `db:"id" json:"-"`
	UserID           uuid.UUID       `db:"user_id" json:"-"`
	TotalAssets      decimal.Decimal `db:"total_assets" json:"total_assets"`
	TotalLiabilities decimal.Decimal `db:"total_liabilities" json:"total_liabilities"`
	NetWorth         decimal.Decimal `db:"net_worth" json:"net_worth"`
	RecordedDate     time.Time       `db:"recorded_date" json:"recorded_date"`
	GrowthPercentage decimal.Decimal `db:"growth_percentage" json:"growth_percentage"`
}

type ListNetworthHistoryRequest struct {
	UserID    uuid.UUID
	Timeframe string `query:"timeframe" validate:"omitempty,oneof=1M 3M 6M YTD 1Y ALL range"` // preset periode
	StartDate string `query:"start_date" validate:"omitempty,date=YYYY-MM-DD"`                // range start date YYYY-MM-DD
	EndDate   string `query:"end_date" validate:"omitempty,date=YYYY-MM-DD"`                  // range end date YYYY-MM-DD
}

type NetworthRepository interface {
	Calculate(ctx context.Context) error
	GetCurrent(ctx context.Context, userId uuid.UUID) (*Networth, error)
	GetNetworthHistory(ctx context.Context, req *ListNetworthHistoryRequest) (*[]Networth, error)
}

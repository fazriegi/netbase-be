package domain

import (
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

package domain

import (
	"github.com/shopspring/decimal"
)

type DashboardCashflowResponse struct {
	Period          string          `json:"period"`
	TotalInflow     decimal.Decimal `json:"total_inflow"`
	TotalOutflow    decimal.Decimal `json:"total_outflow"`
	NetFreeCashflow decimal.Decimal `json:"net_free_cashflow"`
	SavingsRate     decimal.Decimal `json:"savings_rate"`
}

type DashboardCashflowRequest struct {
	Period string `query:"period" validate:"omitempty,date=YYYY-MM"`
}

package domain

import (
	"github.com/google/uuid"
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

type DashboardMilestoneResponse struct {
	ID                 uuid.UUID       `json:"id"`
	Title              string          `json:"title"`
	TargetAmount       decimal.Decimal `json:"target_amount"`
	BaseAmount         decimal.Decimal `json:"base_amount"`
	CurrentNetworth    decimal.Decimal `json:"current_networth"`
	RemainingGap       decimal.Decimal `json:"remaining_gap"`
	ProgressPercentage decimal.Decimal `json:"progress_percentage"`
	IsCompleted        bool            `json:"is_completed"`
}

type DashboardNetworthSummaryResponse struct {
	NetWorth         decimal.Decimal `json:"net_worth"`
	GrowthPercentage decimal.Decimal `json:"growth_percentage"`
	TotalAssets      decimal.Decimal `json:"total_assets"`
	TotalLiabilities decimal.Decimal `json:"total_liabilities"`
	DebtRatio        decimal.Decimal `json:"debt_ratio"`
}

type DashboardNetworthHistoryRequest struct {
	Timeframe string `query:"timeframe" validate:"required,oneof=1M 3M 6M YTD 1Y ALL range"`
	StartDate string `query:"start_date" validate:"omitempty,date=YYYY-MM-DD"`
	EndDate   string `query:"end_date" validate:"omitempty,date=YYYY-MM-DD"`
}

type DashboardNetworthHistoryResponse struct {
	Timeframe string `json:"timeframe"`
	Summary   struct {
		ChangeAmount     decimal.Decimal `json:"change_amount"`
		ChangePercentage decimal.Decimal `json:"change_percentage"`
		IsPositive       bool            `json:"is_positive"`
	} `json:"summary"`
	History []Networth `json:"history"`
}

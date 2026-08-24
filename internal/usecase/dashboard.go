package usecase

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/fazriegi/netbase-be/internal/domain"
	"github.com/fazriegi/netbase-be/pkg"
	"github.com/fazriegi/netbase-be/pkg/constant"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type dashboardUsecase struct {
	log       *log.Logger
	transRepo domain.TransactionRepository
	mlRepo    domain.MilestoneRepository
	nwRepo    domain.NetworthRepository
}

type DashboardUsecase interface {
	GetCashflow(ctx context.Context, period string) (resp pkg.Response)
	GetActiveMilestone(ctx context.Context) (resp pkg.Response)
	GetNetworthSummary(ctx context.Context) (resp pkg.Response)
	GetNetworthHistory(ctx context.Context, req *domain.DashboardNetworthHistoryRequest) (resp pkg.Response)
}

func NewDashboardUsecase(
	log *log.Logger,
	transRepo domain.TransactionRepository,
	mlRepo domain.MilestoneRepository,
	nwRepo domain.NetworthRepository,
) DashboardUsecase {
	return &dashboardUsecase{log, transRepo, mlRepo, nwRepo}
}

func (u *dashboardUsecase) GetCashflow(ctx context.Context, period string) (resp pkg.Response) {
	userId := ctx.Value("user_id").(uuid.UUID)

	realPeriod := period
	if period != "" {
		timeParsed, err := time.Parse("2006-01", period)
		if err != nil {
			return pkg.NewResponse(http.StatusBadRequest, constant.ErrInvalidRequest, nil, nil)
		}
		period = timeParsed.Format("2006-01-02")
	}

	summary, err := u.transRepo.GetSummary(ctx, &domain.ListTransactionRequest{
		UserID:     userId,
		FilterType: "month",
		DateStr:    period,
	})
	if err != nil {
		u.log.Printf("[ERROR] transRepo.GetSummary: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}

	savingsRate := decimal.Zero
	if summary.Income.GreaterThan(decimal.Zero) {
		savingsRate = summary.Net.Div(summary.Income).Mul(decimal.NewFromInt(100)).Round(2)
	}
	dataResponse := domain.DashboardCashflowResponse{
		Period:          realPeriod,
		TotalInflow:     summary.Income,
		TotalOutflow:    summary.Expense,
		NetFreeCashflow: summary.Net,
		SavingsRate:     savingsRate,
	}

	return pkg.NewResponse(http.StatusOK, "Success", dataResponse, nil)
}

func (u *dashboardUsecase) GetActiveMilestone(ctx context.Context) (resp pkg.Response) {
	userId := ctx.Value("user_id").(uuid.UUID)

	milestoneData, err := u.mlRepo.GetCurrent(ctx, &domain.GetMilestone{
		UserId: userId,
	})
	if err != nil {
		u.log.Printf("[ERROR] mlRepo.GetCurrent: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}

	if milestoneData == nil {
		return pkg.NewResponse(http.StatusOK, "No active milestone", nil, nil)
	}

	networthData, err := u.nwRepo.GetCurrent(ctx, userId)
	if err != nil {
		u.log.Printf("[ERROR] nwRepo.GetCurrent: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}

	progressPercentage := decimal.Zero
	if milestoneData.TargetAmount.GreaterThan(decimal.Zero) {
		progressPercentage = networthData.NetWorth.Div(milestoneData.TargetAmount).Mul(decimal.NewFromInt(100)).Round(2)
	}
	dataResponse := domain.DashboardMilestoneResponse{
		ID:                 milestoneData.ID,
		Title:              milestoneData.Title,
		TargetAmount:       milestoneData.TargetAmount,
		BaseAmount:         milestoneData.BaseAmount,
		CurrentNetworth:    networthData.NetWorth,
		RemainingGap:       milestoneData.TargetAmount.Sub(networthData.NetWorth),
		ProgressPercentage: progressPercentage,
		IsCompleted:        milestoneData.TargetAmount.LessThanOrEqual(networthData.NetWorth),
	}

	if dataResponse.IsCompleted != milestoneData.IsCompleted {
		milestoneData.IsCompleted = dataResponse.IsCompleted

		if milestoneData.IsCompleted {
			now := time.Now()
			milestoneData.CompletionDate = &now
		}

		err = u.mlRepo.Update(ctx, milestoneData)
		if err != nil {
			u.log.Printf("[ERROR] mlRepo.Update: %s", err.Error())
			return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
		}
	}

	return pkg.NewResponse(http.StatusOK, "Success", dataResponse, nil)
}

func (u *dashboardUsecase) GetNetworthSummary(ctx context.Context) (resp pkg.Response) {
	userId := ctx.Value("user_id").(uuid.UUID)

	networthData, err := u.nwRepo.GetCurrent(ctx, userId)
	if err != nil {
		u.log.Printf("[ERROR] nwRepo.GetCurrent: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}

	debtRatio := decimal.Zero
	if networthData.TotalAssets.GreaterThan(decimal.Zero) {
		debtRatio = networthData.TotalLiabilities.Div(networthData.TotalAssets).Mul(decimal.NewFromInt(100)).Round(2)
	}

	dataResponse := domain.DashboardNetworthSummaryResponse{
		NetWorth:         networthData.NetWorth,
		GrowthPercentage: networthData.GrowthPercentage,
		TotalAssets:      networthData.TotalAssets,
		TotalLiabilities: networthData.TotalLiabilities,
		DebtRatio:        debtRatio,
	}

	return pkg.NewResponse(http.StatusOK, "Success", dataResponse, nil)
}

func (u *dashboardUsecase) GetNetworthHistory(ctx context.Context, req *domain.DashboardNetworthHistoryRequest) (resp pkg.Response) {
	userId := ctx.Value("user_id").(uuid.UUID)

	histories, err := u.nwRepo.GetNetworthHistory(ctx, &domain.ListNetworthHistoryRequest{
		Timeframe: req.Timeframe,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		UserID:    userId,
	})
	if err != nil {
		u.log.Printf("[ERROR] nwRepo.GetNetworthHistory: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}

	var dataResponse domain.DashboardNetworthHistoryResponse
	dataResponse.Timeframe = req.Timeframe
	dataResponse.History = *histories

	if len(*histories) > 0 {
		first := (*histories)[0].NetWorth
		last := (*histories)[len(*histories)-1].NetWorth

		changeAmount := last.Sub(first)
		changePercentage := decimal.Zero
		if !first.IsZero() {
			changePercentage = changeAmount.Div(first.Abs()).Mul(decimal.NewFromInt(100)).Round(2)
		}

		dataResponse.Summary.ChangeAmount = changeAmount
		dataResponse.Summary.ChangePercentage = changePercentage
		dataResponse.Summary.IsPositive = changeAmount.GreaterThanOrEqual(decimal.Zero)
	} else {
		dataResponse.Summary.ChangeAmount = decimal.Zero
		dataResponse.Summary.ChangePercentage = decimal.Zero
		dataResponse.Summary.IsPositive = true
	}

	return pkg.NewResponse(http.StatusOK, "Success", dataResponse, nil)
}

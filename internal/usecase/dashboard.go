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
}

type DashboardUsecase interface {
	GetCashflow(ctx context.Context, period string) (resp pkg.Response)
}

func NewDashboardUsecase(log *log.Logger, transRepo domain.TransactionRepository) DashboardUsecase {
	return &dashboardUsecase{log, transRepo}
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

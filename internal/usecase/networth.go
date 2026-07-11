package usecase

import (
	"context"
	"log"
	"net/http"

	"github.com/fazriegi/netbase-be/internal/domain"
	"github.com/fazriegi/netbase-be/pkg"
	"github.com/fazriegi/netbase-be/pkg/constant"
	"github.com/google/uuid"
)

type networthUsecase struct {
	log  *log.Logger
	repo domain.NetworthRepository
}

type NetworthUsecase interface {
	GetCurrent(ctx context.Context) (resp pkg.Response)
	CalculateDailyNetworth(ctx context.Context) error
}

func NewNetworthUsecase(log *log.Logger, repo domain.NetworthRepository) NetworthUsecase {
	return &networthUsecase{log, repo}
}

func (u *networthUsecase) GetCurrent(ctx context.Context) (resp pkg.Response) {
	userId := ctx.Value("user_id").(uuid.UUID)

	networth, err := u.repo.GetCurrent(ctx, userId)
	if err != nil {
		if err.Error() != constant.ErrNotFound {
			u.log.Printf("[ERROR] repo.GetCurrent: %s", err.Error())
			return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
		}

		return pkg.NewResponse(http.StatusNotFound, constant.ErrNotFound, nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Success", networth, nil)
}

func (u *networthUsecase) CalculateDailyNetworth(ctx context.Context) error {
	return u.repo.Calculate(ctx)
}


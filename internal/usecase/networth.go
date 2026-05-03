package usecase

import (
	"context"
	"log"
	"net/http"

	"github.com/fazriegi/fintrack-be/internal/repository"
	"github.com/fazriegi/fintrack-be/pkg"
	"github.com/fazriegi/fintrack-be/pkg/constant"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type networthUsecase struct {
	db   *sqlx.DB
	log  *log.Logger
	repo repository.NetworthRepository
}

type NetworthUsecase interface {
	GetCurrent(ctx context.Context) (resp pkg.Response)
}

func NewNetworthUsecase(db *sqlx.DB, log *log.Logger, repo repository.NetworthRepository) NetworthUsecase {
	return &networthUsecase{db, log, repo}
}

func (u *networthUsecase) GetCurrent(ctx context.Context) (resp pkg.Response) {
	userId := ctx.Value("user_id").(uuid.UUID)

	networth, err := u.repo.GetCurrent(ctx, userId, u.db)
	if err != nil {
		if err.Error() != constant.ErrNotFound {
			u.log.Printf("[ERROR] repo.GetCurrent: %s", err.Error())
			return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
		}

		return pkg.NewResponse(http.StatusNotFound, constant.ErrNotFound, nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Success", networth, nil)
}

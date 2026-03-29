package usecase

import (
	"context"
	"log"
	"net/http"

	"github.com/fazriegi/fintrack-be/internal/domain"
	"github.com/fazriegi/fintrack-be/internal/repository"
	"github.com/fazriegi/fintrack-be/pkg"
	"github.com/fazriegi/fintrack-be/pkg/constant"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type assetUsecase struct {
	db   *sqlx.DB
	log  *log.Logger
	repo repository.AssetRepository
}

type AssetUsecase interface {
	ListAsset(ctx context.Context, req *domain.ListAssetRequest) (resp pkg.Response)
}

func NewAssetUsecase(db *sqlx.DB, log *log.Logger, repo repository.AssetRepository) AssetUsecase {
	return &assetUsecase{db, log, repo}
}

func (u *assetUsecase) ListAsset(ctx context.Context, req *domain.ListAssetRequest) (resp pkg.Response) {
	userId := ctx.Value("user_id").(uuid.UUID)
	req.UserId = userId

	assets, err := u.repo.ListAsset(ctx, req, u.db)
	if err != nil {
		u.log.Printf("[ERROR] repo.ListAsset: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Success", &domain.ListAssetResponse{Assets: assets}, nil)
}

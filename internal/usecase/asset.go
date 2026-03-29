package usecase

import (
	"context"
	"log"
	"math"
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
	ListAssetCategory(ctx context.Context) (resp pkg.Response)
}

func NewAssetUsecase(db *sqlx.DB, log *log.Logger, repo repository.AssetRepository) AssetUsecase {
	return &assetUsecase{db, log, repo}
}

func (u *assetUsecase) ListAsset(ctx context.Context, req *domain.ListAssetRequest) (resp pkg.Response) {
	userId := ctx.Value("user_id").(uuid.UUID)
	req.UserId = userId

	assets, total, err := u.repo.ListAsset(ctx, req, u.db)
	if err != nil {
		u.log.Printf("[ERROR] repo.ListAsset: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}

	var paginationMeta pkg.PaginationMeta
	if req.Limit != nil && *req.Limit > 0 {
		limit := int(*req.Limit)
		page := 1

		if req.Page != nil && *req.Page > 0 {
			page = int(*req.Page)
		}

		totalPages := int(math.Ceil(float64(total) / float64(limit)))

		if totalPages > 0 && page > totalPages {
			page = totalPages
		}

		paginationMeta = pkg.PaginationMeta{
			Page:       page,
			Limit:      limit,
			Total:      int(total),
			TotalPages: totalPages,
		}
	}

	return pkg.NewResponse(http.StatusOK, "Success", assets, &paginationMeta)
}

func (u *assetUsecase) ListAssetCategory(ctx context.Context) (resp pkg.Response) {
	userId := ctx.Value("user_id").(uuid.UUID)

	categories, err := u.repo.ListCategory(ctx, userId, u.db)
	if err != nil {
		u.log.Printf("[ERROR] repo.ListAsset: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Success", categories, nil)
}

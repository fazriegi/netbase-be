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
	GetByID(ctx context.Context, id uuid.UUID) (resp pkg.Response)
	Delete(ctx context.Context, id uuid.UUID) (resp pkg.Response)
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

	dataResponse := make([]domain.ListAssetResponse, 0)

	if assets != nil {
		for _, asset := range *assets {
			dataResponse = append(dataResponse, domain.ListAssetResponse{
				ID:           asset.ID,
				Category:     asset.Category,
				Name:         asset.Name,
				CurrentValue: asset.CurrentValue,
				IsActive:     asset.IsActive,
			})
		}
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

	return pkg.NewResponse(http.StatusOK, "Success", dataResponse, &paginationMeta)
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

func (u *assetUsecase) GetByID(ctx context.Context, id uuid.UUID) (resp pkg.Response) {
	userId := ctx.Value("user_id").(uuid.UUID)

	asset, err := u.repo.GetByID(ctx, id, userId, u.db)
	if err != nil {
		if err.Error() != constant.ErrNotFound {
			u.log.Printf("[ERROR] repo.GetByID: %s", err.Error())
			return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
		}

		return pkg.NewResponse(http.StatusNotFound, constant.ErrNotFound, nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Success", asset, nil)
}

func (u *assetUsecase) Delete(ctx context.Context, id uuid.UUID) (resp pkg.Response) {
	userId := ctx.Value("user_id").(uuid.UUID)

	err := u.repo.Delete(ctx, id, userId, u.db)
	if err != nil {
		u.log.Printf("[ERROR] repo.Delete: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Success", nil, nil)
}

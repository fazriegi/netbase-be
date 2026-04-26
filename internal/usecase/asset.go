package usecase

import (
	"context"
	"encoding/json"
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
	Create(ctx context.Context, req *domain.CreateAsset) (resp pkg.Response)
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

	dataResponse := domain.GetAssetByIDResponse{
		ID:           asset.ID,
		CategoryID:   asset.CategoryID,
		Category:     asset.Category,
		CategoryType: asset.CategoryType,
		Name:         asset.Name,
		IsActive:     asset.IsActive,
		CurrentValue: asset.CurrentValue,
	}

	if asset.Details != nil {
		var detailsBytes []byte
		switch v := asset.Details.(type) {
		case []byte:
			detailsBytes = v
		case string:
			detailsBytes = []byte(v)
		}

		if len(detailsBytes) > 0 {
			var parsedDetails any
			if err := json.Unmarshal(detailsBytes, &parsedDetails); err == nil {
				dataResponse.Details = parsedDetails
			}
		}
	}

	return pkg.NewResponse(http.StatusOK, "Success", dataResponse, nil)
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

func (u *assetUsecase) Create(ctx context.Context, req *domain.CreateAsset) (resp pkg.Response) {
	userId := ctx.Value("user_id").(uuid.UUID)
	req.UserId = userId

	var detailsDB any
	if req.Details != nil {
		b, err := json.Marshal(req.Details)
		if err != nil {
			u.log.Printf("[ERROR] json.Marshal req.Details: %s", err.Error())
			return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
		}
		detailsDB = b
	}

	assetDB := &domain.AssetDB{
		UserId:       userId,
		CategoryID:   req.CategoryID,
		Name:         req.Name,
		CurrentValue: *req.CurrentValue,
		Details:      detailsDB,
		IsActive:     *req.IsActive,
	}

	err := u.repo.Insert(ctx, assetDB, u.db)
	if err != nil {
		u.log.Printf("[ERROR] repo.Insert: %s", err.Error())
		return pkg.NewResponse(http.StatusInternalServerError, constant.ErrServer, nil, nil)
	}

	return pkg.NewResponse(http.StatusCreated, "Success", nil, nil)
}

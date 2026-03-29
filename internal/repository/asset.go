package repository

import (
	"context"
	"fmt"

	"github.com/fazriegi/fintrack-be/internal/domain"
	"github.com/fazriegi/fintrack-be/pkg"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type assetRepository struct{}

type AssetRepository interface {
	ListAsset(ctx context.Context, req *domain.ListAssetRequest, db *sqlx.DB) (*[]domain.Asset, error)
	ListCategory(ctx context.Context, userId uuid.UUID, db *sqlx.DB) (*[]domain.Category, error)
}

func NewAssetRepository() AssetRepository {
	return &assetRepository{}
}

func (r *assetRepository) ListAsset(ctx context.Context, req *domain.ListAssetRequest, db *sqlx.DB) (*[]domain.Asset, error) {
	var assets []domain.Asset
	query := `
		SELECT assets.id, assets.user_id, ac.name as category, assets.name, assets.current_value, assets.details, assets.is_active 
		FROM assets 
		join asset_categories ac on ac.id = assets.category_id and ac.user_id = assets.user_id
		WHERE assets.user_id = :user_id
	`

	if req.Name != "" {
		query += ` AND assets.name ILIKE :name`
	}

	if req.Category != "" {
		query += ` AND ac.name ILIKE :category`
	}

	if req.IsActive != nil {
		query += ` AND assets.is_active = :is_active`
	}

	res, err := pkg.SelectWithPagination(ctx, db, query, map[string]interface{}{
		"page":      req.Page,
		"limit":     req.Limit,
		"sort":      req.Sort,
		"user_id":   req.UserId,
		"name":      "%" + req.Name + "%",
		"category":  "%" + req.Category + "%",
		"is_active": req.IsActive,
	})

	if err != nil {
		return nil, err
	}

	for res.Next() {
		var asset domain.Asset
		err := res.StructScan(&asset)
		if err != nil {
			fmt.Printf("[ERROR] scanning asset: %s\n", err.Error())
			continue
		}
		assets = append(assets, asset)
	}

	return &assets, nil
}

func (r *assetRepository) ListCategory(ctx context.Context, userId uuid.UUID, db *sqlx.DB) (*[]domain.Category, error) {
	var categories []domain.Category
	query := `SELECT id, name, base_type FROM asset_categories WHERE user_id = $1 ORDER BY name ASC`
	err := db.SelectContext(ctx, &categories, query, userId)

	return &categories, err
}

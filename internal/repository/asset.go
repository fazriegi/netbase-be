package repository

import (
	"context"
	"fmt"

	"github.com/fazriegi/fintrack-be/internal/domain"
	"github.com/fazriegi/fintrack-be/pkg"
	"github.com/jmoiron/sqlx"
)

type assetRepository struct{}

type AssetRepository interface {
	ListAsset(ctx context.Context, req *domain.ListAssetRequest, db *sqlx.DB) (*[]domain.Asset, error)
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

	res, err := pkg.SelectWithPagination(ctx, db, query, map[string]interface{}{
		"page":    req.Page,
		"limit":   req.Limit,
		"sort":    req.Sort,
		"user_id": req.UserId,
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

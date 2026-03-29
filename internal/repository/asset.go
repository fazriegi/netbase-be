package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/fazriegi/fintrack-be/internal/domain"
	"github.com/fazriegi/fintrack-be/pkg"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type assetRepository struct{}

type AssetRepository interface {
	ListAsset(ctx context.Context, req *domain.ListAssetRequest, db *sqlx.DB) (*[]domain.Asset, int, error)
	ListCategory(ctx context.Context, userId uuid.UUID, db *sqlx.DB) (*[]domain.Category, error)
}

func NewAssetRepository() AssetRepository {
	return &assetRepository{}
}

func (r *assetRepository) ListAsset(ctx context.Context, req *domain.ListAssetRequest, db *sqlx.DB) (*[]domain.Asset, int, error) {
	var assets []domain.Asset
	var total int
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

	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	wg.Add(2)

	go func() {
		defer wg.Done()
		resCount, err := db.NamedQueryContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM (%s) as count_query", query), map[string]interface{}{
			"user_id":   req.UserId,
			"name":      "%" + req.Name + "%",
			"category":  "%" + req.Category + "%",
			"is_active": req.IsActive,
		})

		if err != nil {
			errChan <- fmt.Errorf("error counting data: %v", err)
		}

		defer resCount.Close()

		if resCount.Next() {
			err = resCount.Scan(&total)
			if err != nil {
				errChan <- fmt.Errorf("error scanning count: %v", err)
			}
		}
	}()

	go func() {
		defer wg.Done()
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
			errChan <- fmt.Errorf("error fetching data: %v", err)
		}

		defer res.Close()

		for res.Next() {
			var asset domain.Asset
			err := res.StructScan(&asset)
			if err != nil {
				errChan <- fmt.Errorf("error scanning data: %v", err)
				return
			}
			assets = append(assets, asset)
		}

	}()

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return nil, 0, err
		}
	}

	return &assets, total, nil
}

func (r *assetRepository) ListCategory(ctx context.Context, userId uuid.UUID, db *sqlx.DB) (*[]domain.Category, error) {
	var categories []domain.Category
	query := `SELECT id, name, base_type FROM asset_categories WHERE user_id = $1 ORDER BY name ASC`
	err := db.SelectContext(ctx, &categories, query, userId)

	return &categories, err
}

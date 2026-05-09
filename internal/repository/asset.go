package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/fazriegi/fintrack-be/internal/domain"
	"github.com/fazriegi/fintrack-be/pkg"
	"github.com/fazriegi/fintrack-be/pkg/constant"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

type assetRepository struct{}

type AssetRepository interface {
	ListAsset(ctx context.Context, req *domain.ListAssetRequest, db *sqlx.DB) (*[]domain.Asset, int, error)
	ListCategory(ctx context.Context, userId uuid.UUID, db *sqlx.DB) (*[]domain.Category, error)
	GetByID(ctx context.Context, id, userId uuid.UUID, db *sqlx.DB) (*domain.Asset, error)
	Delete(ctx context.Context, id, userId uuid.UUID, db *sqlx.DB) error
	Insert(ctx context.Context, data *domain.AssetDB, db *sqlx.DB) error
	Update(ctx context.Context, data *domain.AssetDB, db *sqlx.DB) error
	GetTickers(ctx context.Context, db *sqlx.DB) (*[]string, error)
	UpdateStockPrice(ctx context.Context, db *sqlx.DB, ticker string, price decimal.Decimal) error
}

func NewAssetRepository() AssetRepository {
	return &assetRepository{}
}

func (r *assetRepository) ListAsset(ctx context.Context, req *domain.ListAssetRequest, db *sqlx.DB) (*[]domain.Asset, int, error) {
	var assets = make([]domain.Asset, 0)
	var total int
	var defaultSort = "created_at desc"
	query := `
		SELECT assets.id, assets.user_id, ac.name as category, assets.name, assets.current_value, assets.details, assets.is_active,
			assets.created_at
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

	if req.Sort == nil {
		req.Sort = &defaultSort
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
			return
		}

		defer resCount.Close()

		if resCount.Next() {
			err = resCount.Scan(&total)
			if err != nil {
				errChan <- fmt.Errorf("error scanning count: %v", err)
				return
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
			return
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
	var categories = make([]domain.Category, 0)
	query := `SELECT id, name, base_type FROM asset_categories WHERE user_id = $1 ORDER BY name ASC`
	err := db.SelectContext(ctx, &categories, query, userId)

	return &categories, err
}

func (r *assetRepository) GetByID(ctx context.Context, id, userId uuid.UUID, db *sqlx.DB) (*domain.Asset, error) {
	var asset domain.Asset
	query := `
		SELECT assets.id, assets.user_id, assets.category_id, assets.name, assets.current_value, assets.details, assets.is_active, ac."name" as category, 
			ac.base_type as category_type
		FROM assets 
		JOIN asset_categories ac ON assets.user_id = ac.user_id AND assets.category_id = ac.id
		WHERE assets.id = $1 AND assets.user_id = $2`
	err := db.GetContext(ctx, &asset, query, id, userId)
	if err == sql.ErrNoRows {
		return nil, errors.New(constant.ErrNotFound)
	}

	return &asset, err
}

func (r *assetRepository) Delete(ctx context.Context, id, userId uuid.UUID, db *sqlx.DB) error {
	query := `DELETE FROM assets WHERE id = $1 AND user_id = $2`
	_, err := db.ExecContext(ctx, query, id, userId)

	return err
}

func (r *assetRepository) Insert(ctx context.Context, data *domain.AssetDB, db *sqlx.DB) error {
	query := `INSERT INTO assets (user_id, category_id, name, current_value, details, is_active) VALUES (:user_id, :category_id, :name, :current_value, :details, :is_active)`
	_, err := db.NamedExecContext(ctx, query, data)

	return err
}

func (r *assetRepository) Update(ctx context.Context, data *domain.AssetDB, db *sqlx.DB) error {
	query := `UPDATE assets SET name = :name, category_id = :category_id, current_value = :current_value, details = :details, is_active = :is_active, updated_at = now() WHERE id = :id AND user_id = :user_id`
	_, err := db.NamedExecContext(ctx, query, data)

	return err
}

func (r *assetRepository) GetTickers(ctx context.Context, db *sqlx.DB) (*[]string, error) {
	var tickers = make([]string, 0)
	query := `
		SELECT distinct assets.details->>'ticker_symbol' AS ticker_symbol
		FROM assets 
		JOIN asset_categories ac ON assets.user_id = ac.user_id AND assets.category_id = ac.id
		WHERE ac.base_type = 'investment'
			AND ac.name = 'Stock'
			AND assets.is_active = true
	`
	err := db.SelectContext(ctx, &tickers, query)

	return &tickers, err
}

func (r *assetRepository) UpdateStockPrice(ctx context.Context, db *sqlx.DB, ticker string, price decimal.Decimal) error {
	query := `
		UPDATE assets 
		SET current_value = (details->>'quantity')::decimal * $1, updated_at = now() 
		WHERE details->>'ticker_symbol' = $2
	`
	_, err := db.ExecContext(ctx, query, price, ticker)

	return err
}

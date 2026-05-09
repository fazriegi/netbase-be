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
)

type liabilityRepository struct{}

type LiabilityRepository interface {
	ListCategory(ctx context.Context, userId uuid.UUID, db *sqlx.DB) (*[]domain.Category, error)
	List(ctx context.Context, req *domain.ListLiabilityRequest, db *sqlx.DB) (*[]domain.Liability, int, error)
	GetByID(ctx context.Context, id, userId uuid.UUID, db *sqlx.DB) (*domain.Liability, error)
	Delete(ctx context.Context, id, userId uuid.UUID, db *sqlx.DB) error
	Insert(ctx context.Context, data *domain.LiabilityDB, db *sqlx.DB) error
	Update(ctx context.Context, data *domain.LiabilityDB, db *sqlx.DB) error
}

func NewLiabilityRepository() LiabilityRepository {
	return &liabilityRepository{}
}

func (r *liabilityRepository) ListCategory(ctx context.Context, userId uuid.UUID, db *sqlx.DB) (*[]domain.Category, error) {
	var categories = make([]domain.Category, 0)
	query := `SELECT id, name, base_type FROM liability_categories WHERE user_id = $1 ORDER BY name ASC`
	err := db.SelectContext(ctx, &categories, query, userId)

	return &categories, err
}

func (r *liabilityRepository) List(ctx context.Context, req *domain.ListLiabilityRequest, db *sqlx.DB) (*[]domain.Liability, int, error) {
	var liabilities = make([]domain.Liability, 0)
	var total int
	query := `
		SELECT liabilities.id, liabilities.user_id, lc.name as category, liabilities.name, liabilities.remaining_balance, liabilities.details 
		FROM liabilities 
		join liability_categories lc on lc.id = liabilities.category_id and lc.user_id = liabilities.user_id
		WHERE liabilities.user_id = :user_id
	`

	if req.Name != "" {
		query += ` AND liabilities.name ILIKE :name`
	}

	if req.Category != "" {
		query += ` AND lc.name ILIKE :category`
	}

	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	wg.Add(2)

	go func() {
		defer wg.Done()
		resCount, err := db.NamedQueryContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM (%s) as count_query", query), map[string]interface{}{
			"user_id":  req.UserId,
			"name":     "%" + req.Name + "%",
			"category": "%" + req.Category + "%",
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
			"page":     req.Page,
			"limit":    req.Limit,
			"sort":     req.Sort,
			"user_id":  req.UserId,
			"name":     "%" + req.Name + "%",
			"category": "%" + req.Category + "%",
		})

		if err != nil {
			errChan <- fmt.Errorf("error fetching data: %v", err)
			return
		}

		defer res.Close()

		for res.Next() {
			var liability domain.Liability
			err := res.StructScan(&liability)
			if err != nil {
				errChan <- fmt.Errorf("error scanning data: %v", err)
				return
			}
			liabilities = append(liabilities, liability)
		}

	}()

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return nil, 0, err
		}
	}

	return &liabilities, total, nil
}

func (r *liabilityRepository) Delete(ctx context.Context, id, userId uuid.UUID, db *sqlx.DB) error {
	query := `DELETE FROM liabilities WHERE id = $1 AND user_id = $2`
	_, err := db.ExecContext(ctx, query, id, userId)

	return err
}

func (r *liabilityRepository) Insert(ctx context.Context, data *domain.LiabilityDB, db *sqlx.DB) error {
	query := `INSERT INTO liabilities (user_id, category_id, name, principal_amount, remaining_balance, details) VALUES (:user_id, :category_id, :name, :principal_amount, :remaining_balance, :details)`
	_, err := db.NamedExecContext(ctx, query, data)

	return err
}

func (r *liabilityRepository) Update(ctx context.Context, data *domain.LiabilityDB, db *sqlx.DB) error {
	query := `UPDATE liabilities SET name = :name, category_id = :category_id, principal_amount = :principal_amount, remaining_balance = :remaining_balance, details = :details, updated_at = now() WHERE id = :id AND user_id = :user_id`
	_, err := db.NamedExecContext(ctx, query, data)

	return err
}

func (r *liabilityRepository) GetByID(ctx context.Context, id, userId uuid.UUID, db *sqlx.DB) (*domain.Liability, error) {
	var liability domain.Liability
	query := `
		SELECT liabilities.id, liabilities.user_id, liabilities.category_id, liabilities.name, 
			liabilities.principal_amount, liabilities.remaining_balance, liabilities.details, 
			lc.name as category, lc.base_type as category_type
		FROM liabilities
		JOIN liability_categories lc ON liabilities.user_id = lc.user_id AND liabilities.category_id = lc.id
		WHERE liabilities.id = $1 AND liabilities.user_id = $2`
	err := db.GetContext(ctx, &liability, query, id, userId)
	if err == sql.ErrNoRows {
		return nil, errors.New(constant.ErrNotFound)
	}

	return &liability, err
}

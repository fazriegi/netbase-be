package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/fazriegi/fintrack-be/internal/domain"
	"github.com/fazriegi/fintrack-be/pkg"
	"github.com/fazriegi/fintrack-be/pkg/constant"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type transactionRepository struct {
	db *sqlx.DB
}

func NewTransactionRepository(db *sqlx.DB) domain.TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) ListCategory(ctx context.Context, req *domain.ListCategoryRequest) (*[]domain.Category, error) {
	db := getQueryer(ctx, r.db)
	var categories = make([]domain.Category, 0)
	query := `SELECT id, name, base_type FROM transaction_categories WHERE user_id = $1`

	args := []interface{}{req.UserID}

	if req.BaseType != "" {
		query += ` AND base_type = $2`
		args = append(args, req.BaseType)
	}

	if req.Search != "" {
		query += ` AND name ILIKE $3`
		args = append(args, "%"+req.Search+"%")
	}

	query += ` ORDER BY name ASC`

	err := db.SelectContext(ctx, &categories, query, args...)

	return &categories, err
}

func (r *transactionRepository) GetCategoryByID(ctx context.Context, id, userID uuid.UUID) (*domain.Category, error) {
	db := getQueryer(ctx, r.db)
	var category domain.Category
	query := `SELECT id, name, base_type FROM transaction_categories WHERE id = $1 AND user_id = $2`
	err := db.GetContext(ctx, &category, query, id, userID)
	if err == sql.ErrNoRows {
		return nil, errors.New(constant.ErrNotFound)
	}

	return &category, err
}

func (r *transactionRepository) InsertCategory(ctx context.Context, category *domain.Category) error {
	db := getQueryer(ctx, r.db)
	query := `INSERT INTO transaction_categories (user_id, name, base_type) VALUES (:user_id, :name, :base_type) ON CONFLICT (user_id, name, base_type) DO NOTHING`
	_, err := db.NamedExecContext(ctx, query, category)

	return err
}

func (r *transactionRepository) DeleteCategory(ctx context.Context, id, userID uuid.UUID) error {
	db := getQueryer(ctx, r.db)
	query := `DELETE FROM transaction_categories WHERE id = $1 AND user_id = $2`
	_, err := db.ExecContext(ctx, query, id, userID)

	if err != nil && strings.Contains(err.Error(), "violates RESTRICT setting of foreign key constraint") {
		return errors.New("violates foreign key constraint")
	}

	return err
}

func (r *transactionRepository) transactionFilter(req *domain.ListTransactionRequest) string {
	var query string
	if req.CategoryName != "" {
		query += ` AND tc.name ILIKE :category_name`
	}

	if req.Notes != "" {
		query += ` AND transactions.notes ILIKE :notes`
	}

	switch req.FilterType {
	case "week":
		query += ` AND DATE_TRUNC('week', transactions.transaction_date) = DATE_TRUNC('week', CAST(:ref_date AS date))`
	case "month":
		query += ` AND DATE_TRUNC('month', transactions.transaction_date) = DATE_TRUNC('month', CAST(:ref_date AS date))`
	case "year":
		query += ` AND DATE_TRUNC('year', transactions.transaction_date) = DATE_TRUNC('year', CAST(:ref_date AS date))`
	case "range":
		query += ` AND DATE_TRUNC('day', transactions.transaction_date) BETWEEN DATE_TRUNC('day', CAST(:start_date AS date)) AND DATE_TRUNC('day', CAST(:end_date AS date))`
	}

	return query
}

func (r *transactionRepository) List(ctx context.Context, req *domain.ListTransactionRequest) (*[]domain.Transaction, int, error) {
	db := getQueryer(ctx, r.db)
	var transactions = make([]domain.Transaction, 0)
	var total int
	var defaultSort = "transaction_date desc, created_at desc"

	query := `
		SELECT 
			transactions.id, 
			transactions.user_id, 
			transactions.asset_id, 
			assets.name as asset_name,
			transactions.liability_id, 
			liabilities.name as liability_name,
			transactions.category_id, 
			tc.name as category_name, 
			tc.base_type as category_type,
			transactions.amount, 
			transactions.transaction_date, 
			transactions.notes,
			transactions.created_at
		FROM transactions 
		JOIN transaction_categories tc ON tc.id = transactions.category_id AND tc.user_id = transactions.user_id
		LEFT JOIN assets ON assets.id = transactions.asset_id AND assets.user_id = transactions.user_id
		LEFT JOIN liabilities ON liabilities.id = transactions.liability_id AND liabilities.user_id = transactions.user_id
		WHERE transactions.user_id = :user_id
	`

	var refDate string
	if req.DateStr != "" {
		refDate = req.DateStr
	} else {
		refDate = time.Now().Format("2006-01-02")
	}

	query += r.transactionFilter(req)

	if req.Sort == nil {
		req.Sort = &defaultSort
	}

	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	wg.Add(2)

	arg := map[string]interface{}{
		"user_id":       req.UserID,
		"category_name": "%" + req.CategoryName + "%",
		"notes":         "%" + req.Notes + "%",
		"ref_date":      refDate,
		"start_date":    req.StartDateStr,
		"end_date":      req.EndDateStr,
	}
	go func() {
		defer wg.Done()
		resCount, err := db.NamedQueryContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM (%s) as count_query", query), arg)

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

		arg["page"] = req.Page
		arg["limit"] = req.Limit
		arg["sort"] = req.Sort
		res, err := pkg.SelectWithPagination(ctx, db, query, arg)

		if err != nil {
			errChan <- fmt.Errorf("error fetching data: %v", err)
			return
		}

		defer res.Close()

		for res.Next() {
			var tx domain.Transaction
			err := res.StructScan(&tx)
			if err != nil {
				errChan <- fmt.Errorf("error scanning data: %v", err)
				return
			}
			transactions = append(transactions, tx)
		}
	}()

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return nil, 0, err
		}
	}

	return &transactions, total, nil
}

func (r *transactionRepository) GetSummary(ctx context.Context, req *domain.ListTransactionRequest) (*domain.TransactionSummary, error) {
	db := getQueryer(ctx, r.db)

	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN tc.base_type = 'income' THEN transactions.amount ELSE 0 END), 0) as income,
			COALESCE(SUM(CASE WHEN tc.base_type = 'expense' THEN transactions.amount ELSE 0 END), 0) as expense
		FROM transactions 
		JOIN transaction_categories tc ON tc.id = transactions.category_id AND tc.user_id = transactions.user_id
		WHERE transactions.user_id = :user_id
	`

	var refDate string
	if req.DateStr != "" {
		refDate = req.DateStr
	} else {
		refDate = time.Now().Format("2006-01-02")
	}

	query += r.transactionFilter(req)

	rows, err := db.NamedQueryContext(ctx, query, map[string]interface{}{
		"user_id":       req.UserID,
		"category_name": "%" + req.CategoryName + "%",
		"notes":         "%" + req.Notes + "%",
		"ref_date":      refDate,
		"start_date":    req.StartDateStr,
		"end_date":      req.EndDateStr,
	})
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summary domain.TransactionSummary
	if rows.Next() {
		if err := rows.Scan(&summary.Income, &summary.Expense); err != nil {
			return nil, err
		}
	}
	summary.Net = summary.Income.Sub(summary.Expense)

	return &summary, nil
}

func (r *transactionRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.Transaction, error) {
	db := getQueryer(ctx, r.db)
	var tx domain.Transaction
	query := `
		SELECT 
			transactions.id, 
			transactions.user_id, 
			transactions.asset_id, 
			assets.name as asset_name,
			transactions.liability_id, 
			liabilities.name as liability_name,
			transactions.category_id, 
			tc.name as category_name, 
			tc.base_type as category_type,
			transactions.amount, 
			transactions.transaction_date, 
			transactions.notes,
			transactions.created_at
		FROM transactions 
		JOIN transaction_categories tc ON tc.id = transactions.category_id AND tc.user_id = transactions.user_id
		LEFT JOIN assets ON assets.id = transactions.asset_id AND assets.user_id = transactions.user_id
		LEFT JOIN liabilities ON liabilities.id = transactions.liability_id AND liabilities.user_id = transactions.user_id
		WHERE transactions.id = $1 AND transactions.user_id = $2
	`
	err := db.GetContext(ctx, &tx, query, id, userID)
	if err == sql.ErrNoRows {
		return nil, errors.New(constant.ErrNotFound)
	}

	return &tx, err
}

func (r *transactionRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	db := getQueryer(ctx, r.db)
	query := `DELETE FROM transactions WHERE id = $1 AND user_id = $2`
	_, err := db.ExecContext(ctx, query, id, userID)

	return err
}

func (r *transactionRepository) Insert(ctx context.Context, data *domain.TransactionDB) error {
	db := getQueryer(ctx, r.db)
	query := `
		INSERT INTO transactions (user_id, asset_id, liability_id, category_id, amount, transaction_date, notes) 
		VALUES (:user_id, :asset_id, :liability_id, :category_id, :amount, :transaction_date, :notes)
	`
	_, err := db.NamedExecContext(ctx, query, data)

	return err
}

func (r *transactionRepository) Update(ctx context.Context, data *domain.TransactionDB) error {
	db := getQueryer(ctx, r.db)
	query := `
		UPDATE transactions 
		SET asset_id = :asset_id, liability_id = :liability_id, category_id = :category_id, amount = :amount, transaction_date = :transaction_date, notes = :notes, updated_at = now() 
		WHERE id = :id AND user_id = :user_id
	`
	_, err := db.NamedExecContext(ctx, query, data)

	return err
}

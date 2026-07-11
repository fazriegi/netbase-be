package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/fazriegi/netbase-be/internal/domain"
	"github.com/fazriegi/netbase-be/pkg/constant"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type networthRepository struct {
	db *sqlx.DB
}

func NewNetworthRepository(db *sqlx.DB) domain.NetworthRepository {
	return &networthRepository{db: db}
}

func (r *networthRepository) Calculate(ctx context.Context) error {
	db := getQueryer(ctx, r.db)
	query := `
		WITH AssetSummary AS (
			SELECT user_id, COALESCE(SUM(current_value), 0) AS total_assets
			FROM assets 
			WHERE is_active = TRUE 
			GROUP BY user_id
		),
		LiabilitySummary AS (
			SELECT user_id, COALESCE(SUM(remaining_balance), 0) AS total_liabilities
			FROM liabilities 
			WHERE remaining_balance > 0 
			GROUP BY user_id
		)
		INSERT INTO net_worth_histories (user_id, total_assets, total_liabilities, recorded_date)
		SELECT 
			U.id, 
			COALESCE(A.total_assets, 0), 
			COALESCE(L.total_liabilities, 0),
			CURRENT_DATE
		FROM users U
		LEFT JOIN AssetSummary A ON U.id = A.user_id
		LEFT JOIN LiabilitySummary L ON U.id = L.user_id
		ON CONFLICT (user_id, recorded_date) 
		DO UPDATE SET 
			total_assets = EXCLUDED.total_assets,
			total_liabilities = EXCLUDED.total_liabilities,
			updated_at = NOW();`

	_, err := db.ExecContext(ctx, query)

	return err
}

func (r *networthRepository) GetCurrent(ctx context.Context, userId uuid.UUID) (*domain.Networth, error) {
	db := getQueryer(ctx, r.db)
	var networth domain.Networth
	query := `
		WITH realtime_assets AS (
			SELECT COALESCE(SUM(current_value), 0) AS total_assets
			FROM assets
			WHERE user_id = $1 AND is_active = TRUE
		),
		realtime_liabilities AS (
			SELECT COALESCE(SUM(remaining_balance), 0) AS total_liabilities
			FROM liabilities
			WHERE user_id = $1 AND remaining_balance > 0
		),
		last_month_snapshot AS (
			SELECT net_worth
			FROM net_worth_histories
			WHERE user_id = $1 
			AND recorded_date < DATE_TRUNC('month', CURRENT_DATE)
			ORDER BY recorded_date DESC
			LIMIT 1
		)
		SELECT 
			(ra.total_assets - rl.total_liabilities) AS net_worth,
			ra.total_assets,
			rl.total_liabilities,
			CASE 
				WHEN lms.net_worth IS NULL OR lms.net_worth = 0 THEN 0
				ELSE (((ra.total_assets - rl.total_liabilities) - lms.net_worth) / lms.net_worth) * 100
			END AS growth_percentage
		FROM realtime_assets ra
		CROSS JOIN realtime_liabilities rl
		LEFT JOIN last_month_snapshot lms ON TRUE;`
	err := db.GetContext(ctx, &networth, query, userId)
	if err == sql.ErrNoRows {
		return nil, errors.New(constant.ErrNotFound)
	}

	return &networth, err
}

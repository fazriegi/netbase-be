package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/fazriegi/netbase-be/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type milestoneRepository struct {
	db *sqlx.DB
}

func NewMilestoneRepository(db *sqlx.DB) domain.MilestoneRepository {
	return &milestoneRepository{db: db}
}

func (r *milestoneRepository) Insert(ctx context.Context, data *domain.Milestone) error {
	db := getQueryer(ctx, r.db)
	query := `
		INSERT INTO milestones (user_id, title, base_amount, target_amount)
		VALUES (:user_id, :title, :base_amount, :target_amount)
	`

	_, err := db.NamedExecContext(ctx, query, data)

	return err
}

func (r *milestoneRepository) GetCurrent(ctx context.Context, data *domain.GetMilestone) (*domain.Milestone, error) {
	db := getQueryer(ctx, r.db)
	var milestone domain.Milestone

	query := `
		SELECT id, user_id, title, base_amount, target_amount, is_completed, completion_date
		FROM milestones
		WHERE user_id = $1
	`
	args := []interface{}{data.UserId}

	if data.Title != "" {
		query += fmt.Sprintf(` AND title = $%d`, len(args)+1)
		args = append(args, data.Title)
	}

	query += ` ORDER BY created_at DESC LIMIT 1`

	err := db.GetContext(ctx, &milestone, query, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return &milestone, err
}

func (r *milestoneRepository) Update(ctx context.Context, data *domain.Milestone) error {
	db := getQueryer(ctx, r.db)
	query := `
		UPDATE milestones
		SET base_amount = :base_amount, target_amount = :target_amount, updated_at = now(), 
		    is_completed = :is_completed, completion_date = :completion_date
		WHERE user_id = :user_id AND title = :title
	`

	_, err := db.NamedExecContext(ctx, query, data)

	return err
}

func (r *milestoneRepository) Delete(ctx context.Context, id uuid.UUID) error {
	db := getQueryer(ctx, r.db)
	query := `DELETE FROM milestones WHERE id = $1`
	_, err := db.ExecContext(ctx, query, id)

	return err
}

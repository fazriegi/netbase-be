package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// type MilestoneDB struct {
// 	ID             uuid.UUID       `db:"id"`
// 	UserId         uuid.UUID       `db:"user_id"`
// 	Title          string          `db:"title"`
// 	BaseAmount     decimal.Decimal `db:"base_amount"`
// 	TargetAmount   decimal.Decimal `db:"target_amount"`
// 	IsCompleted    bool            `db:"is_completed"`
// 	CompletionDate *time.Time      `db:"completion_date"`
// 	CreatedAt      time.Time       `db:"created_at"`
// 	UpdatedAt      time.Time       `db:"updated_at"`
// }

type Milestone struct {
	ID             uuid.UUID       `db:"id" json:"id"`
	UserId         uuid.UUID       `db:"user_id" json:"-"`
	Title          string          `db:"title" json:"title"`
	BaseAmount     decimal.Decimal `db:"base_amount" json:"base_amount"`
	TargetAmount   decimal.Decimal `db:"target_amount" json:"target_amount"`
	IsCompleted    bool            `db:"is_completed" json:"is_completed"`
	CompletionDate *time.Time      `db:"completion_date" json:"completion_date"`
	CreatedAt      time.Time       `db:"created_at" json:"-"`
	UpdatedAt      time.Time       `db:"updated_at" json:"-"`
}

type CreateMilestoneRequest struct {
	Title        string          `json:"title" validate:"required,max=255"`
	TargetAmount decimal.Decimal `json:"target_amount" validate:"required,decimal"`
}

type GetMilestone struct {
	UserId uuid.UUID `json:"user_id"`
	Title  string    `json:"title"`
}

type MilestoneRepository interface {
	Insert(ctx context.Context, data *Milestone) error
	GetCurrent(ctx context.Context, data *GetMilestone) (*Milestone, error)
	Update(ctx context.Context, data *Milestone) error
	Delete(ctx context.Context, id uuid.UUID) error
}

package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/fazriegi/fintrack-be/internal/domain"
	"github.com/fazriegi/fintrack-be/pkg/constant"
	"github.com/google/uuid"

	"github.com/jmoiron/sqlx"
)

type userRepo struct{}

type UserRepository interface {
	Create(ctx context.Context, user *domain.User, tx *sqlx.Tx) (uuid.UUID, error)
	GetByEmail(ctx context.Context, email string, db *sqlx.DB) (*domain.User, error)
	GetByID(ctx context.Context, userId uuid.UUID, db *sqlx.DB) (*domain.User, error)
	CheckRefreshToken(ctx context.Context, userId uuid.UUID, refreshToken string, db *sqlx.DB) (exp time.Time, err error)
	InsertRefreshToken(ctx context.Context, data domain.RefreshToken, tx *sqlx.Tx) error
	SeedDefaultCategories(ctx context.Context, tx *sqlx.Tx, userID uuid.UUID) error
	RevokeRefreshToken(ctx context.Context, userID uuid.UUID, refreshToken string, tx *sqlx.Tx) error
	RemoveExpiredToken(ctx context.Context, userID *uuid.UUID, db *sqlx.DB) error
}

func NewUserRepository() UserRepository {
	return &userRepo{}
}

func (r *userRepo) Create(ctx context.Context, user *domain.User, tx *sqlx.Tx) (uuid.UUID, error) {
	query := `INSERT INTO users (email, password, full_name) VALUES ($1, $2, $3) RETURNING id`
	var userId uuid.UUID
	err := tx.QueryRowContext(ctx, query, user.Email, user.Password, user.FullName).Scan(&userId)
	return userId, err
}

func (r *userRepo) GetByEmail(ctx context.Context, email string, db *sqlx.DB) (*domain.User, error) {
	var user domain.User
	query := `SELECT id, email, password, full_name FROM users WHERE email = $1`
	err := db.GetContext(ctx, &user, query, email)
	if err == sql.ErrNoRows {
		return nil, errors.New(constant.ErrUserNotFound)
	}
	return &user, err
}

func (r *userRepo) GetByID(ctx context.Context, userId uuid.UUID, db *sqlx.DB) (*domain.User, error) {
	var user domain.User
	query := `SELECT id, email, password, full_name FROM users WHERE id = $1`
	err := db.GetContext(ctx, &user, query, userId)
	if err == sql.ErrNoRows {
		return nil, errors.New(constant.ErrUserNotFound)
	}

	return &user, err
}

func (r *userRepo) CheckRefreshToken(ctx context.Context, userId uuid.UUID, refreshToken string, db *sqlx.DB) (exp time.Time, err error) {
	query := `
		SELECT expires_at
		FROM refresh_tokens 
		WHERE user_id = $1
			AND token = $2
			AND is_revoked = false
			AND expires_at > now()
	`

	err = db.QueryRowContext(ctx, query, userId, refreshToken).Scan(&exp)

	if err != nil {
		if err == sql.ErrNoRows {
			return exp, errors.New(constant.ErrNotFound)
		}

		return exp, err
	}

	return exp, nil
}

func (r *userRepo) InsertRefreshToken(ctx context.Context, data domain.RefreshToken, tx *sqlx.Tx) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token, expires_at, device_info, ip_address) 
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := tx.ExecContext(
		ctx,
		query,
		data.UserID,
		data.Token,
		data.ExpiresAt,
		data.DeviceInfo,
		data.IPAddress,
	)

	return err
}

func (r *userRepo) RevokeRefreshToken(ctx context.Context, userID uuid.UUID, refreshToken string, tx *sqlx.Tx) error {
	query := `
		UPDATE refresh_tokens 
		SET is_revoked = TRUE
		WHERE user_id = $1
			AND token = $2
	`

	_, err := tx.ExecContext(ctx, query, userID, refreshToken)
	if err != nil {
		return err
	}

	return nil
}

func (r *userRepo) RemoveExpiredToken(ctx context.Context, userID *uuid.UUID, db *sqlx.DB) error {
	query := `
		DELETE FROM refresh_tokens
		WHERE expires_at < now()
	`

	if userID != nil {
		query += ` AND user_id = $1`
	}

	param := []interface{}{}
	if userID != nil {
		param = append(param, *userID)
	}

	_, err := db.ExecContext(ctx, query, param...)

	return err
}

func (r *userRepo) SeedDefaultCategories(ctx context.Context, tx *sqlx.Tx, userID uuid.UUID) error {
	defaultAssets := []domain.Category{
		{UserID: userID, Name: "Cash", BaseType: "liquid"},
		{UserID: userID, Name: "Savings", BaseType: "liquid"},
		{UserID: userID, Name: "Bond", BaseType: "investment"},
		{UserID: userID, Name: "Crypto", BaseType: "investment"},
		{UserID: userID, Name: "Gold", BaseType: "physical"},
		{UserID: userID, Name: "Mutual Fund", BaseType: "investment"},
		{UserID: userID, Name: "Stock", BaseType: "investment"},
		{UserID: userID, Name: "Electronics", BaseType: "physical"},
		{UserID: userID, Name: "Property", BaseType: "physical"},
		{UserID: userID, Name: "Vehicle", BaseType: "physical"},
	}

	defaultLiabilities := []domain.Category{
		{UserID: userID, Name: "Credit Card", BaseType: "short_term"},
		{UserID: userID, Name: "Paylater", BaseType: "short_term"},
		{UserID: userID, Name: "Personal Loan", BaseType: "short_term"},
		{UserID: userID, Name: "Loan", BaseType: "long_term"},
	}

	defaultTransactions := []domain.Category{
		{UserID: userID, Name: "Salary", BaseType: "income"},
		{UserID: userID, Name: "Dividend", BaseType: "income"},
		{UserID: userID, Name: "Freelance", BaseType: "income"},
		{UserID: userID, Name: "Investment", BaseType: "income"},
		{UserID: userID, Name: "Savings", BaseType: "income"},
		{UserID: userID, Name: "Other", BaseType: "income"},
		{UserID: userID, Name: "Bills", BaseType: "expense"},
		{UserID: userID, Name: "Entertainment", BaseType: "expense"},
		{UserID: userID, Name: "Health", BaseType: "expense"},
		{UserID: userID, Name: "Social", BaseType: "expense"},
		{UserID: userID, Name: "Top-up", BaseType: "expense"},
		{UserID: userID, Name: "Food & Dining", BaseType: "expense"},
		{UserID: userID, Name: "Transportation", BaseType: "expense"},
		{UserID: userID, Name: "Fuel & Vehicle Maintenance", BaseType: "expense"},
		{UserID: userID, Name: "Subscriptions", BaseType: "expense"},
		{UserID: userID, Name: "Investment", BaseType: "expense"},
		{UserID: userID, Name: "Savings", BaseType: "expense"},
		{UserID: userID, Name: "Shopping", BaseType: "expense"},
		{UserID: userID, Name: "Other", BaseType: "expense"},
	}

	queryAsset := `
		INSERT INTO asset_categories (user_id, name, base_type) 
		VALUES (:user_id, :name, :base_type)`
	if _, err := tx.NamedExecContext(ctx, queryAsset, defaultAssets); err != nil {
		return err
	}

	queryLiability := `
		INSERT INTO liability_categories (user_id, name, base_type) 
		VALUES (:user_id, :name, :base_type)`
	if _, err := tx.NamedExecContext(ctx, queryLiability, defaultLiabilities); err != nil {
		return err
	}

	queryTransaction := `
		INSERT INTO transaction_categories (user_id, name, base_type) 
		VALUES (:user_id, :name, :base_type)`
	if _, err := tx.NamedExecContext(ctx, queryTransaction, defaultTransactions); err != nil {
		return err
	}

	return nil
}

package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `db:"id" json:"id"`
	FullName     string    `db:"full_name" json:"full_name"`
	Username     string    `db:"username" json:"username"`
	Email        string    `db:"email" json:"email"`
	Password     string    `db:"password" json:"-"`
	RefreshToken string    `db:"refresh_token" json:"-"`
}

type RegisterRequest struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"omitempty,email"`
	Password string `json:"password" validate:"required,password"`
	FullName string `json:"full_name" validate:"required"`
}

type LoginRequest struct {
	Username   string `json:"username" validate:"required"`
	Password   string `json:"password" validate:"required"`
	RemoteAddr string `json:"-"`
}

type RefreshToken struct {
	UserID     uuid.UUID
	Token      string
	ExpiresAt  time.Time
	DeviceInfo string
	IPAddress  string
}

type UserRepository interface {
	Create(ctx context.Context, user *User) (uuid.UUID, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, userId uuid.UUID) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	CheckRefreshToken(ctx context.Context, userId uuid.UUID, refreshToken string) (exp time.Time, err error)
	InsertRefreshToken(ctx context.Context, data RefreshToken) error
	SeedDefaultCategories(ctx context.Context, userID uuid.UUID) error
	RevokeRefreshToken(ctx context.Context, userID uuid.UUID, refreshToken string) error
	RemoveExpiredToken(ctx context.Context, userID *uuid.UUID) error
}

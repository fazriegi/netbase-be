package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `db:"id" json:"id"`
	FullName     string    `db:"full_name" json:"full_name"`
	Email        string    `db:"email" json:"email"`
	Password     string    `db:"password" json:"-"`
	RefreshToken string    `db:"refresh_token" json:"-"`
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,password"`
	FullName string `json:"full_name" validate:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,password"`
}

type RefreshToken struct {
	UserID     uuid.UUID
	Token      string
	ExpiresAt  time.Time
	DeviceInfo string
	IPAddress  string
}

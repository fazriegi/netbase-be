package cron

import (
	"log"

	"github.com/fazriegi/fintrack-be/internal/repository"
	"github.com/jmoiron/sqlx"
)

func Start(db *sqlx.DB, logger *log.Logger) {
	// Refresh Token
	userRepo := repository.NewUserRepository()
	go func() {
		RefreshTokenCleanup(db, userRepo, logger)
	}()
}

package cron

import (
	"context"
	"log"
	"time"

	"github.com/fazriegi/fintrack-be/internal/repository"
	"github.com/go-co-op/gocron"
	"github.com/jmoiron/sqlx"
)

func RefreshTokenCleanup(db *sqlx.DB, userRepo repository.UserRepository, appLogger *log.Logger) {
	s := gocron.NewScheduler(time.Local)

	s.Every(1).Day().Do(func() {
		err := userRepo.RemoveExpiredToken(context.Background(), nil, db)
		if err != nil {
			appLogger.Printf("Error occurred while removing expired tokens: %v", err)
		}

		appLogger.Println("Refresh token cleanup executed at:", time.Now())
	})

	s.StartAsync()

	// Prevent exit
	select {}
}

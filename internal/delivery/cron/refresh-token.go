package cron

import (
	"context"
	"log"
	"time"

	"github.com/fazriegi/netbase-be/internal/usecase"
	"github.com/go-co-op/gocron"
)

func RefreshTokenCleanup(userUC usecase.UserUsecase, appLogger *log.Logger) {
	s := gocron.NewScheduler(time.Local)

	_, err := s.Every(1).Day().At("03:00").Do(func() {
		safeExecute(appLogger, "RefreshTokenCleanup", func() {
			appLogger.Println("Starting expired refresh token cleanup...")

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
			defer cancel()

			err := userUC.CleanupExpiredTokens(ctx)
			if err != nil {
				appLogger.Printf("ERROR: Failed to remove expired tokens: %v", err)
				return
			}

			appLogger.Printf("SUCCESS: Expired refresh tokens cleaned up at: %s", time.Now().Format("15:04:05"))
		})
	})

	if err != nil {
		appLogger.Fatalf("Failed to schedule cleanup job: %v", err)
	}

	s.StartAsync()
	appLogger.Println("Token cleanup scheduler is active.")
}

package cron

import (
	"context"
	"log"
	"time"

	"github.com/fazriegi/netbase-be/internal/usecase"
	"github.com/go-co-op/gocron"
)

func NetworthCalculate(networthUC usecase.NetworthUsecase, appLogger *log.Logger) {
	s := gocron.NewScheduler(time.Local)

	_, err := s.Every(1).Day().At("23:59").Do(func() {
		safeExecute(appLogger, "NetworthCalculate", func() {
			appLogger.Println("Starting scheduled net worth calculation...")

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			err := networthUC.CalculateDailyNetworth(ctx)
			if err != nil {
				appLogger.Printf("ERROR: Failed to calculate net worth: %v", err)
				return
			}

			appLogger.Printf("SUCCESS: Net worth recorded at %s", time.Now().Format("2006-01-02 15:04:05"))
		})
	})

	if err != nil {
		appLogger.Fatalf("Failed to schedule job: %v", err)
	}

	s.StartAsync()

	appLogger.Println("Net worth scheduler is active.")
}

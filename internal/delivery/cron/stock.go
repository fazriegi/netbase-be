package cron

import (
	"context"
	"log"
	"time"

	"github.com/fazriegi/netbase-be/internal/usecase"
	"github.com/go-co-op/gocron"
)

func UpdateStockPrice(assetUC usecase.AssetUsecase, appLogger *log.Logger) {
	s := gocron.NewScheduler(time.Local)

	_, err := s.Every(1).Day().At("17:00").Do(func() {
		safeExecute(appLogger, "UpdateStockPrice", func() {
			today := time.Now().Weekday()
			if today == time.Saturday || today == time.Sunday {
				appLogger.Println("Skipping scheduled stock price update: It's the weekend.")
				return
			}

			appLogger.Println("Starting scheduled stock price update...")

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
			defer cancel()

			err := assetUC.UpdateStockPrices(ctx)
			if err != nil {
				appLogger.Printf("ERROR: Failed to update stock prices: %v", err)
				return
			}

			appLogger.Printf("SUCCESS: Stock prices updated at %s", time.Now().Format("2006-01-02 15:04:05"))
		})
	})

	if err != nil {
		appLogger.Fatalf("Failed to schedule job: %v", err)
	}

	s.StartAsync()

	appLogger.Println("Stock price update scheduler is active.")
}

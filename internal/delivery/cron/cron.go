package cron

import (
	"log"
	"os"

	"github.com/fazriegi/netbase-be/internal/infrastructure/yahoo"
	"github.com/fazriegi/netbase-be/internal/repository"
	"github.com/fazriegi/netbase-be/internal/usecase"
	"github.com/jmoiron/sqlx"
)

func Start(db *sqlx.DB, logger *log.Logger) {
	txManager := repository.NewTransactionManager(db)

	// Refresh Token
	userRepo := repository.NewUserRepository(db)
	userUC := usecase.NewUserUsecase(logger, userRepo, txManager)
	go func() {
		RefreshTokenCleanup(userUC, logger)
	}()

	// Networth
	networthRepo := repository.NewNetworthRepository(db)
	networthUC := usecase.NewNetworthUsecase(logger, networthRepo)
	go func() {
		NetworthCalculate(networthUC, logger)
	}()

	// Stock
	yahooProvider := yahoo.NewYahooProvider(os.Getenv("RAPID_API_KEY"))
	assetRepo := repository.NewAssetRepository(db)
	assetUC := usecase.NewAssetUsecase(logger, assetRepo, yahooProvider)
	go func() {
		UpdateStockPrice(assetUC, logger)
	}()
}

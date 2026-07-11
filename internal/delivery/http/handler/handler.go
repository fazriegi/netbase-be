package handler

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/fazriegi/netbase-be/internal/delivery/http/middleware"
	"github.com/fazriegi/netbase-be/internal/infrastructure/yahoo"
	"github.com/fazriegi/netbase-be/internal/repository"
	"github.com/fazriegi/netbase-be/internal/usecase"
	"github.com/jmoiron/sqlx"

	"github.com/rs/cors"
)

func New(db *sqlx.DB, logger *log.Logger) http.Handler {
	txManager := repository.NewTransactionManager(db)

	// USER
	userRepo := repository.NewUserRepository(db)
	authUC := usecase.NewUserUsecase(logger, userRepo, txManager)

	// ASSET
	yahooProvider := yahoo.NewYahooProvider(os.Getenv("RAPID_API_KEY"))
	assetRepo := repository.NewAssetRepository(db)
	assetUC := usecase.NewAssetUsecase(logger, assetRepo, yahooProvider)

	// LIABILITY
	liabilityRepo := repository.NewLiabilityRepository(db)
	liabilityUC := usecase.NewLiabilityUsecase(logger, liabilityRepo)

	// NETWORTH
	networthRepo := repository.NewNetworthRepository(db)
	networthUC := usecase.NewNetworthUsecase(logger, networthRepo)

	// TRANSACTION
	transactionRepo := repository.NewTransactionRepository(db)
	transactionUC := usecase.NewTransactionUsecase(logger, transactionRepo, txManager, assetRepo, liabilityRepo)

	mux := http.NewServeMux()

	NewUserHandler(mux, authUC, logger)
	NewAssetHandler(mux, assetUC, logger)
	NewLiabilityHandler(mux, liabilityUC, logger)
	NewNetworthHandler(mux, networthUC, logger)
	NewTransactionHandler(mux, transactionUC, logger)

	origin := os.Getenv("ALLOWED_ORIGIN")
	if origin == "" {
		origin = "*"
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   strings.Split(origin, ","),
		AllowedMethods:   []string{"POST", "GET", "OPTIONS", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Requested-With"},
		AllowCredentials: true,
	})

	recovery := middleware.Recovery(logger)
	return recovery(c.Handler(mux))
}

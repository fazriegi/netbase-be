package handler

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/fazriegi/fintrack-be/internal/repository"
	"github.com/fazriegi/fintrack-be/internal/usecase"
	"github.com/jmoiron/sqlx"

	"github.com/rs/cors"
)

func New(db *sqlx.DB, logger *log.Logger) http.Handler {
	// USER
	userRepo := repository.NewUserRepository()
	authUC := usecase.NewUserUsecase(db, logger, userRepo)

	// ASSET
	assetRepo := repository.NewAssetRepository()
	assetUC := usecase.NewAssetUsecase(db, logger, assetRepo)

	// LIABILITY
	liabilityRepo := repository.NewLiabilityRepository()
	liabilityUC := usecase.NewLiabilityUsecase(db, logger, liabilityRepo)

	// NETWORTH
	networthRepo := repository.NewNetworthRepository()
	networthUC := usecase.NewNetworthUsecase(db, logger, networthRepo)

	mux := http.NewServeMux()

	NewUserHandler(mux, authUC, logger)
	NewAssetHandler(mux, assetUC, logger)
	NewLiabilityHandler(mux, liabilityUC, logger)
	NewNetworthHandler(mux, networthUC, logger)

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

	return c.Handler(mux)
}

package handler

import (
	"log"
	"net/http"

	"github.com/fazriegi/fintrack-be/internal/delivery/http/middleware"
	"github.com/fazriegi/fintrack-be/internal/repository"
	"github.com/fazriegi/fintrack-be/internal/usecase"
	"github.com/jmoiron/sqlx"
)

func New(db *sqlx.DB, logger *log.Logger) http.Handler {
	// USER
	userRepo := repository.NewUserRepository()
	authUC := usecase.NewUserUsecase(db, logger, userRepo)

	mux := http.NewServeMux()

	NewUserHandler(mux, authUC, logger)

	return middleware.MiddlewareCORS(mux)
}

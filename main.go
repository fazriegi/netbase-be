package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/fazriegi/netbase-be/internal/delivery/cron"
	"github.com/fazriegi/netbase-be/internal/delivery/http/handler"
	"github.com/fazriegi/netbase-be/pkg/database"
	"github.com/fazriegi/netbase-be/pkg/logger"
	"github.com/fazriegi/netbase-be/pkg/token"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: Error loading .env file:", err)
	}

	logFilename := os.Getenv("LOG_FILE")
	if logFilename == "" {
		logFilename = "app.log"
	}

	appLogger, logOutput := logger.SetupLogger(logFilename)

	defer func() {
		if err := logOutput.Close(); err != nil {
			log.Printf("Error closing log file: %v", err)
		}
	}()

	appLogger.Println("Starting application...")

	if err := token.InitSecretKey(); err != nil {
		appLogger.Fatalf("Security configuration error: %v\n", err)
	}

	db := database.ConnectPostgres(appLogger)
	defer db.Close()

	cron.Start(db, appLogger)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	handler := handler.New(db, appLogger)

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	appLogger.Println("Server running on port", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		appLogger.Fatalf("could not start server: %v\n", err)
	}
}

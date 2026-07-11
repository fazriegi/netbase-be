package main

import (
	"log"
	"net/http"
	"os"

	"github.com/fazriegi/fintrack-be/internal/delivery/cron"
	"github.com/fazriegi/fintrack-be/internal/delivery/http/handler"
	"github.com/fazriegi/fintrack-be/pkg/database"
	"github.com/fazriegi/fintrack-be/pkg/logger"
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

	db := database.ConnectPostgres(appLogger)
	defer db.Close()

	cron.Start(db, appLogger)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	handler := handler.New(db, appLogger)

	appLogger.Println("Server running on port", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		appLogger.Fatalf("could not start server: %v\n", err)
	}
}

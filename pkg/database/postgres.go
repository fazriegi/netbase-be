package database

import (
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func ConnectPostgres(appLogger *log.Logger) *sqlx.DB {
	appLogger.Println("Connecting to PostgreSQL...")

	dbDSN := os.Getenv("DATABASE_URL")
	if dbDSN == "" {
		appLogger.Fatalln("Failed to connect to PostgreSQL: DATABASE_URL environment variable is not set")
	}

	db, err := sqlx.Connect("postgres", dbDSN)
	if err != nil {
		appLogger.Fatalln("Failed to connect to PostgreSQL:", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	return db
}

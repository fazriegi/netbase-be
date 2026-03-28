package database

import (
	"log"

	"github.com/jmoiron/sqlx"
)

var db *sqlx.DB

func Get() *sqlx.DB {
	if db == nil {
		log.Fatal("database connection is not initialized")
	}
	return db
}

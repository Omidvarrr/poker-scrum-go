package database

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

// RunMigrations runs all pending migrations on the database file
func RunMigrations(dbPath string) error {
	// Open database with foreign keys enabled
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		return err
	}
	defer db.Close()

	// Enable foreign keys
	if _, err = db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return err
	}

	// Set up migrations
	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}

	// Run migrations
	return goose.Up(db, "migrations")
}

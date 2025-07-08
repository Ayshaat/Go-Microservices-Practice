package db

import (
	"database/sql"
	_ "embed"
	"fmt"
	"github.com/pressly/goose/v3"
	"log"

	_ "github.com/lib/pq"
)

func ConnectDB(connStr string, migrationFiles string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	if err := RunMigrations(db, migrationFiles); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

func RunMigrations(db *sql.DB, migrationFiles string) error {
	if err := goose.Up(db, migrationFiles); err != nil {
		return fmt.Errorf("migrations: %w", err)
	}

	log.Println("Migrations applied successfully.")

	return nil
}

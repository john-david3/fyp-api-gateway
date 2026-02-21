package db

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewDatabase(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		slog.Error("error opening database connection:", "error", err)
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}

func StartDB(db *sql.DB) error {
	path := filepath.Join("/var/lib/init.sql")

	content, err := os.ReadFile(path)
	if err != nil {
		slog.Error("error reading database file", "error", err)
		return err
	}

	if _, err = db.Exec(string(content)); err != nil {
		slog.Error("error executing database statement", "error", err)
		return err
	}

	return nil
}

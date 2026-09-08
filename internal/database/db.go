package database

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	migrations "le-grimoire/db"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

func Open(dbPath string) (*sql.DB, error) {
	// Create Directory if not exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("creating db dir: %w", err)
	}

	dsn := fmt.Sprintf("%s?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=ON", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Ping verifies the connection is actually valid
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("database unreachable: %w", err)
	}

	// Initialize db migrations
	goose.SetBaseFS(migrations.MigrationsFS)
	if err := goose.SetDialect("sqlite"); err != nil {
		return nil, fmt.Errorf("failed to set goose dialect: %w", err)
	}

	slog.Info("running database migrations...")
	if err := goose.Up(db, "migrations"); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}
	slog.Info("database migrations completed")

	return db, nil
}

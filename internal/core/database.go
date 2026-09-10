package core

import (
	"database/sql"
	"fmt"
	"le-grimoire/internal/config"
	"le-grimoire/internal/database"
	"log/slog"
	"path/filepath"
)

// NewDatabase opens the SQLite database and runs migrations.
// The returned shutdown func closes the connection; callers must call it
// before the process exits (after the HTTP server has stopped accepting requests).
func NewDatabase(cfg *config.Config, logger *slog.Logger) (*sql.DB, func(), error) {
	dbPath := filepath.Join(cfg.AppDataDir, cfg.Database.Name) + ".db"
	db, err := database.Open(dbPath)
	if err != nil {
		return nil, nil, fmt.Errorf("open database: %w", err)
	}

	shutdown := func() {
		logger.Info("Closing database connection...")
		if err := db.Close(); err != nil {
			logger.Error("error closing database", "error", err)
		}
		logger.Info("Database connection closed")
	}
	return db, shutdown, nil
}

package core

import (
	"le-grimoire/internal/database"
	"path/filepath"
)

func (a *App) InitDatabase() {
	dbPath := filepath.Join(a.Config.AppDataDir, a.Config.Database.Name) + ".db"
	db, err := database.Open(dbPath)
	if err != nil {
		a.Logger.Error("failed to open database", "err", err)
		panic(err)
	}
	a.Database = db
}

func (a *App) ShutdownDatabase() {
	if a.Database != nil {
		return
	}

	a.Logger.Info("Closing database connection...")
	if err := a.Database.Close(); err != nil {
		a.Logger.Error("error closing database", "error", err)
	}

	a.Logger.Info("Database connection closed")
}

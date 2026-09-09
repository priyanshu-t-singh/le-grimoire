package core

import (
	"fmt"
	"le-grimoire/internal/config"
	"le-grimoire/internal/util"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/lmittmann/tint"
)

// NewLogger creates a dual console+file structured logger based on cfg.
// The returned shutdown func closes the underlying log file; callers must
// call it before the process exits.
func NewLogger(cfg *config.Config) (*slog.Logger, func(), error) {
	consoleHandler := tint.NewTextHandler(os.Stdout, &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: time.TimeOnly,
	})

	if cfg.Logs.Dir != "" {
		if err := os.MkdirAll(cfg.Logs.Dir, 0755); err != nil {
			return nil, nil, fmt.Errorf("create logs dir: %w", err)
		}
	}

	logFileName := fmt.Sprintf("le-grimoire-%s.log", time.Now().Format("2006-01-02_15-04-05"))
	logFilePath := filepath.Join(cfg.Logs.Dir, logFileName)
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, nil, fmt.Errorf("open log file: %w", err)
	}

	jsonHandler := slog.NewJSONHandler(logFile, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	})

	multi := util.NewMultiHandler(consoleHandler, jsonHandler)
	wrapped := &util.ContextHandler{Handler: multi}
	logger := slog.New(wrapped)
	slog.SetDefault(logger)

	shutdown := func() {
		slog.Info("Shutting down logger...")
		if err := logFile.Close(); err != nil {
			slog.Error("failed to close log file", "error", err)
		}
	}
	return logger, shutdown, nil
}

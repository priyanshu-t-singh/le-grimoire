package core

import (
	"le-grimoire/internal/config"
	"le-grimoire/internal/device"
	"log/slog"
)

// All construction now happens in server/server.go.
type App struct {
	Config           *config.Config
	Logger           *slog.Logger
	DeviceRepository *device.Repository

	// Shutdown callbacks populated by the composition root.
	ShutdownDB     func()
	ShutdownLogger func()
}

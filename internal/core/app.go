package core

import (
	"database/sql"
	"le-grimoire/internal/config"
	"le-grimoire/internal/constants"
	"le-grimoire/internal/device"
	"le-grimoire/internal/library"
	"le-grimoire/internal/provider"
	"le-grimoire/internal/render"
	"le-grimoire/internal/state"
	"le-grimoire/internal/util"
	"log/slog"
)

// TODO: set up a proper dependency injection framework for this app,
// so that we can easily swap out implementations of the repositories, renderers, etc.
type App struct {
	// Config
	Config   *config.Config
	Logger   *slog.Logger
	Database *sql.DB
	Version  string

	// Repository
	DeviceRepository *device.Repository

	// Dependencies Interfaces
	BookRepository library.BookProvider

	StateMachine *state.Machine
	Renderer     *render.Renderer
	FrameCache   *render.FrameCache

	// Shutdown
	ShutdownLogger func()
}

func NewApp(configOpts *config.ConfigOptions) *App {
	logger := util.NewLogger()
	logger.Info("Initializing application...")

	cfg, err := config.NewConfig(configOpts, logger)
	if err != nil {
		logger.Error("failed to load config", "error", err)
		return nil
	}

	bookProvider, err := provider.NewBookProvider(*cfg)
	if err != nil {
		logger.Error("failed to initialize book provider", "error", err)
		return nil
	}

	return &App{
		Config:         cfg,
		Logger:         logger,
		Database:       nil,
		Version:        constants.Version,
		BookRepository: bookProvider,
	}
}

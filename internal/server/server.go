package server

import (
	"context"
	"le-grimoire/internal/config"
	"le-grimoire/internal/core"
	"le-grimoire/internal/device"
	"le-grimoire/internal/handlers"
	"le-grimoire/internal/provider"
	"le-grimoire/internal/render"
	"le-grimoire/internal/state"
	"le-grimoire/internal/util"
	"log"

	"github.com/joho/godotenv"
)

func StartServer(flags config.LeGrimoireFlags) {
	startApp(flags)
}

func startApp(flags config.LeGrimoireFlags) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	bootstrapLogger := util.NewLogger()
	bootstrapLogger.Info("Initializing application...")

	// Config
	cfg, err := config.NewConfig(&config.ConfigOptions{Flags: flags}, bootstrapLogger)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Structured logger (console + file; replaces bootstrap)
	logger, shutdownLogger, err := core.NewLogger(cfg)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	// Book provider
	books, err := provider.NewBookProvider(*cfg)
	if err != nil {
		logger.Error("failed to initialize book provider", "error", err)
		log.Fatal(err)
	}
	if err := books.Connect(context.Background()); err != nil {
		logger.Error("failed to connect to book provider", "error", err)
		log.Fatal(err)
	}
	logger.Info("Connected to book provider", "provider", cfg.BookBackend)

	// Database
	db, shutdownDB, err := core.NewDatabase(cfg, logger)
	if err != nil {
		logger.Error("failed to open database", "error", err)
		log.Fatal(err)
	}

	// Concrete deps (order-independent past this point)
	deviceRepo := device.NewDeviceRepository(db)
	stateMachine := state.NewMachine(books, logger)
	renderer := render.NewRenderer(context.Background(), cfg.ChromeRemoteURL, cfg.ChromePath)
	frameCache := render.NewFrameCache()

	// App carry-bag — only what RunHTTPServer / graceful-shutdown needs
	app := &core.App{
		Config:           cfg,
		Logger:           logger,
		DeviceRepository: deviceRepo,
		ShutdownDB:       shutdownDB,
		ShutdownLogger:   shutdownLogger,
	}

	// HTTP router + handler
	router := core.NewHTTPServer()
	h := &handlers.Handler{
		Books:    books,
		Renderer: renderer,
		Devices:  deviceRepo,
		States:   stateMachine,
		Cache:    frameCache,
		Log:      logger,
	}
	handlers.InitRoutes(h, router)

	// Run (blocks until SIGINT/SIGTERM, then shuts down)
	core.RunHTTPServer(app, router)
}

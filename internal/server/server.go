package server

import (
	"le-grimoire/internal/config"
	"le-grimoire/internal/core"
	"le-grimoire/internal/handlers"
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

	// Create the app instance
	app := core.NewApp(&config.ConfigOptions{Flags: flags})
	app.InitLogging()
	app.InitDatabase()
	app.InitRepositories()

	// Initialize the HTTP server
	router := core.NewHTTPServer()

	// Wire concrete deps into the handler via narrow interfaces.
	h := &handlers.Handler{
		Books:    app.BookRepository,
		Renderer: app.Renderer,
		Devices:  app.DeviceRepository,
		States:   app.StateMachine,
		Cache:    app.FrameCache,
		Log:      app.Logger,
	}

	// Initialize the routes
	handlers.InitRoutes(h, router)

	// Run the server with graceful shutdown
	core.RunHTTPServer(app, router)
}

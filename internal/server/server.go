package server

import (
	"le-grimoire/internal/config"
	"le-grimoire/internal/core"
	"le-grimoire/internal/handlers"
	"log"

	"github.com/joho/godotenv"
)

func StartServer() {
	startApp()
}

func startApp() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Get the flags
	flags := config.GetLeGrimoireFlags()

	// Create the app instance
	app := core.NewApp(&config.ConfigOptions{Flags: flags})
	app.InitLogging()
	app.InitDatabase()
	app.InitRepositories()

	// Initialize the HTTP server
	router := core.NewHTTPServer()

	// Initialize the routes
	handlers.InitRoutes(app, router)

	// Run the server with graceful shutdown
	core.RunHTTPServer(app, router)
}

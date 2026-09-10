package core

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"le-grimoire/internal/middleware"
)

func NewHTTPServer() *http.ServeMux {
	return http.NewServeMux()
}

func RunHTTPServer(app *App, router *http.ServeMux) {
	server := &http.Server{
		Addr: app.Config.GetServerAddr(),
		Handler: middleware.CreateStack(
			middleware.Logging,
			middleware.AllowCors,
			middleware.RateLimiter,
			middleware.DeviceAuth(app.DeviceRepository),
		)(router),
	}

	runServerWithGracefulShutdown(app, server)
}

func runServerWithGracefulShutdown(app *App, server *http.Server) {
	serverAddr := app.Config.GetServerAddr()
	app.Logger.Info(
		"Starting HTTP server",
		slog.String("address", serverAddr),
		slog.String("version", app.Config.Version),
	)

	go func() {
		app.Logger.Info("Server is running at " + app.Config.GetServerURI())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.Logger.Error("server failed to serve", "error", err)
			panic(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		app.Logger.Error("server forced to shutdown", "error", err)
	}

	app.ShutdownDB()
	app.ShutdownLogger()
	slog.Info("Shutdown complete")
}

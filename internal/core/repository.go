package core

import (
	"context"
	"log"

	"le-grimoire/internal/device"
	"le-grimoire/internal/render"
	"le-grimoire/internal/state"
)

// TODO: set up a proper dependency injection for this app
func (a *App) InitRepositories() {
	a.StateMachine = state.NewMachine(a.BookRepository, a.Logger)
	a.Renderer = render.NewRenderer(context.Background(), a.Config.ChromeRemoteURL, a.Config.ChromePath)
	a.FrameCache = render.NewFrameCache()

	// one-time auth at boot
	if err := a.BookRepository.Connect(context.Background()); err != nil {
		log.Fatal(err)
	} else {
		a.Logger.Info("Successfully connected to book provider", "provider", a.Config.BookBackend)
	}

	// Initialize Device Repository
	a.DeviceRepository = device.NewDeviceRepository(a.Database)
}

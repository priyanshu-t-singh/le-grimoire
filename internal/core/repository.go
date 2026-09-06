package core

import (
	"context"
	"log"
	"time"

	"le-grimoire/internal/device"
	"le-grimoire/internal/kavita"
	"le-grimoire/internal/render"
	"le-grimoire/internal/state"
	"le-grimoire/internal/util"
)

// TODO: set up a proper dependency injection framework for this app
func (a *App) InitRepositories() {
	// Initialize Kavita Repository
	kavitaClient := kavita.NewClient(kavita.Config{
		BaseURL: util.GetKavitaBaseURL(),
		APIKey:  util.GetKavitaAPIKey(),
	})
	kavitaRepo := kavita.NewRepository(kavitaClient)

	a.KavitaRepository = kavitaRepo

	authCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	a.StateMachine = state.NewMachine(kavitaRepo, a.Logger)
	a.Renderer = render.NewRenderer(context.Background())
	a.FrameCache = render.NewFrameCache()

	if err := kavitaRepo.Authenticate(authCtx); err != nil {
		a.Logger.Warn("Initial Kavita auth failed (will retry reactively on requests)", "error", err)
	} else {
		a.Logger.Info("Successfully authenticated with Kavita server")
	}

	// one-time auth at boot
	if err := a.BookRepository.Connect(context.Background()); err != nil {
		log.Fatal(err)
	}

	// Initialize Device Repository
	a.DeviceRepository = device.NewDeviceRepository(a.Database)
}

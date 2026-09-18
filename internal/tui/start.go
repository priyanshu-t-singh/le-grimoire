package tui

import (
	"context"
	"fmt"
	"le-grimoire/internal/config"
	"le-grimoire/internal/provider"
	"le-grimoire/internal/util"
	"log"

	tea "charm.land/bubbletea/v2"

	"github.com/joho/godotenv"
)

func Start(ctx context.Context, flags config.LeGrimoireFlags) error {
	// Load .env file if present (same pattern as the HTTP server).
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	bootstrapLogger := util.NewLogger()

	// Load config
	cfg, err := config.NewConfig(&config.ConfigOptions{Flags: flags}, bootstrapLogger)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	books, err := provider.NewBookProvider(*cfg)
	if err != nil {
		return fmt.Errorf("init book provider: %w", err)
	}
	if err := books.Connect(ctx); err != nil {
		return fmt.Errorf("connect to book provider: %w", err)
	}

	// Run the Bubble Tea program.
	m := NewModel(ctx, books)
	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("tui run: %w", err)
	}
	return nil
}

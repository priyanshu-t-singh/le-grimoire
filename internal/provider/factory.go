package provider

import (
	"fmt"
	"le-grimoire/internal/config"
	"le-grimoire/internal/library"
	"le-grimoire/internal/provider/kavita"
	"le-grimoire/internal/provider/localfs"
)

func NewBookProvider(cfg config.Config) (library.BookProvider, error) {
	switch cfg.BookBackend {
	case "kavita":
		return kavita.New(cfg.KavitaURL, cfg.KavitaAPIKey), nil
	case "localfs":
		return localfs.New(cfg.AppDataDir), nil
	default:
		return nil, fmt.Errorf("unknown backend %q", cfg.BookBackend)
	}
}

package config

import (
	"flag"
	"fmt"
	"le-grimoire/internal/constants"
	"os"
	"runtime"
	"strings"
)

type (
	LeGrimoireFlags struct {
		DataDir string
		Host    string
		Port    int
	}
)

func GetLeGrimoireFlags() LeGrimoireFlags {
	flags := LeGrimoireFlags{}

	flag.Usage = func() {
		fmt.Printf("Le Grimoire server\n\n")
		if runtime.GOOS == "windows" {
			fmt.Printf("Usage: le-grimoire.exe [flags]\n\n")
		} else {
			fmt.Printf("Usage: le-grimoire [flags]\n\n")
		}
		fmt.Printf("Flags:\n")
		fmt.Printf("  --datadir string              directory that contains all le-grimoire data\n")
		fmt.Printf("  --host string                 host address to bind to (default: 127.0.0.1)\n")
		fmt.Printf("  --port int                    port to bind to (default: 8321)\n")
		fmt.Printf("  --version                     show the version of le-grimoire\n")
		fmt.Printf("  -h                            show this help message\n")
	}

	flag.StringVar(&flags.DataDir, "datadir", "", "Directory that contains all le-grimoire data")
	flag.StringVar(&flags.Host, "host", "", "Host address to bind to")
	flag.IntVar(&flags.Port, "port", 0, "Port to bind to")
	flag.Bool("version", false, "Show the version of le-grimoire")

	flag.Parse()

	if flag.Lookup("version").Value.(flag.Getter).Get().(bool) {
		fmt.Printf("Le Grimoire %s (%s/%s)\n", constants.Version, runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}

	flags.DataDir = strings.TrimSpace(flags.DataDir)
	flags.Host = strings.TrimSpace(flags.Host)

	return flags
}

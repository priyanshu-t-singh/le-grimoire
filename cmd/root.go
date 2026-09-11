package cmd

import (
	"context"
	"fmt"
	"le-grimoire/internal/config"
	"le-grimoire/internal/constants"
	"le-grimoire/internal/tui"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

var AppFlags config.LeGrimoireFlags

var rootCmd = &cobra.Command{
	Use:     "le-grimoire",
	Short:   "A simple e-ink book reader server",
	Long:    `Le Grimoire is a simple e-ink book reader server that serves books to e-ink devices.`,
	Version: fmt.Sprintf("%s (%s/%s)", constants.Version, runtime.GOOS, runtime.GOARCH),
	RunE: func(cmd *cobra.Command, args []string) error {
		AppFlags.Clean()
		return runTerminalApp(cmd.Context())
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&AppFlags.DataDir, "datadir", "", "directory that contains all le-grimoire data")
}

// runTerminalApp is the default action when no subcommand is given.
// It launches the full-screen terminal book reader.
func runTerminalApp(ctx context.Context) error {
	return tui.Start(ctx, AppFlags)
}

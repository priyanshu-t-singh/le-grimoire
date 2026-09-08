package cmd

import (
	"context"
	"fmt"
	"le-grimoire/internal/config"
	"le-grimoire/internal/constants"
	"le-grimoire/internal/core"
	"log/slog"
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

		slog.Info("data dir", "data_dir", AppFlags.DataDir)
		ctx := cmd.Context()
		return runTerminalApp(ctx, nil)
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

func runTerminalApp(ctx context.Context, a *core.App) error {
	// TODO: Implement a terminal-based UI for the application.
	slog.WarnContext(ctx, "Terminal UI is not yet implemented. Please use the 'serve' command to start the server.")
	return nil
}

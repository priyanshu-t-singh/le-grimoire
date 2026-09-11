package cmd

import (
	"context"
	"le-grimoire/internal/tui"

	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the terminal book reader",
	Long:  `Start an interactive full-screen terminal UI for browsing and reading books.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		AppFlags.Clean()
		return runTUI(cmd.Context())
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}

func runTUI(ctx context.Context) error {
	return tui.Start(ctx, AppFlags)
}

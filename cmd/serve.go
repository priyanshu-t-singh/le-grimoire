package cmd

import (
	"context"
	"le-grimoire/internal/server"

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the e-ink book reader server",
	Long:  `Start the e-ink book reader server to serve books to e-ink devices.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		AppFlags.Clean()
		ctx := cmd.Context()
		return runServer(ctx)
	},
}

func init() {
	serveCmd.Flags().StringVar(&AppFlags.Host, "host", "", "host address to bind to")
	serveCmd.Flags().IntVar(&AppFlags.Port, "port", 0, "port to bind to")

	rootCmd.AddCommand(serveCmd)
}

func runServer(ctx context.Context) error {
	server.StartServer(ctx, AppFlags)
	return nil
}

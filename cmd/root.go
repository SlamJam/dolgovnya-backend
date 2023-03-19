package cmd

import (
	"os"

	"github.com/SlamJam/dolgovnya-backend/internal/bootstrap/fxhttp"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var rootCmd = &cobra.Command{
	Use:   "dolgovnya",
	Short: "Dolgovnya debt",
	Long:  `A Fast and Flexible debt management`,
	Run: func(cmd *cobra.Command, args []string) {
		fx.New(
			fxhttp.Module,
			fx.Invoke(func(fxhttp.HTTPServer) {}),
			fx.Invoke(func(fxhttp.ConnectServer) {}),
		).Run()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		rootCmd.PrintErr(err)
		os.Exit(1)
	}
}

package cmd

import (
	"fmt"
	"os"

	"github.com/SlamJam/dolgovnya-backend/internal/bootstrap/fxapp"
	"github.com/SlamJam/dolgovnya-backend/internal/bootstrap/fxhttp"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var verbose bool

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}

var rootCmd = &cobra.Command{
	SilenceErrors: true,
	SilenceUsage:  true,
	Use:           "dolgovnya",
	Short:         "A Fast and Flexible debt management",
	Run: func(cmd *cobra.Command, args []string) {
		// Start main app
		fxapp.NewApp(
			fxapp.Module,
			// Запускаем те сервисы, которые составляю наше приложение
			fx.Invoke(func(fxhttp.HTTPServer) {}),
			fx.Invoke(func(fxhttp.ConnectServer) {}),
		).Run()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "Got error: %+v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "Got error: %v\n", err)
		}

		os.Exit(1)
	}
}

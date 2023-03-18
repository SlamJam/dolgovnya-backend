package cmd

import (
	"context"
	"database/sql"

	"github.com/SlamJam/dolgovnya-backend/internal/app/config"
	"github.com/SlamJam/dolgovnya-backend/internal/bootstrap/fxapp"
	"github.com/SlamJam/dolgovnya-backend/migrations"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

func init() {
	rootCmd.AddCommand(migrationCmd)
}

var migrationCmd = &cobra.Command{
	Use:   "migration",
	Short: "Database migrations",
	Long:  `Migrate database`,
	Run: func(cmd *cobra.Command, args []string) {
		var cfg config.Config
		app := fx.New(
			fxapp.Module,
			fx.Populate(&cfg),
		)

		if err := app.Start(context.Background()); err != nil {
			panic(err)
		}
		defer app.Stop(context.Background())

		goose.SetBaseFS(migrations.FS)

		if err := goose.SetDialect("postgres"); err != nil {
			panic(err)
		}

		// setup database
		db, err := sql.Open("pgx", cfg.DSN)
		if err != nil {
			panic(err)
		}

		if err := goose.Up(db, "."); err != nil {
			panic(err)
		}
	},
}

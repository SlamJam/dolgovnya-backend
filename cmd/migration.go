package cmd

import (
	"context"
	"database/sql"

	"github.com/SlamJam/dolgovnya-backend/internal/app/config"
	"github.com/SlamJam/dolgovnya-backend/internal/bootstrap/fxapp"
	"github.com/SlamJam/dolgovnya-backend/migrations"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

func init() {
	migrationCmd.AddCommand(migrationUpCmd)
	migrationCmd.AddCommand(migrationUpByOneCmd)
	migrationCmd.AddCommand(migrationVersionCmd)
	migrationCmd.AddCommand(migrationStatusCmd)

	rootCmd.AddCommand(migrationCmd)
}

var migrationCmd = &cobra.Command{
	Use:   "migration",
	Short: "Database migrations",
	Long:  `Migrate database`,
}

func prepareGoose() error {
	goose.SetBaseFS(migrations.FS)

	// TODO:
	// goose.SetLogger()

	return goose.SetDialect("postgres")
}

func prepareDB(ctx context.Context) (_ *sql.DB, _ func() error, resultError error) {
	var cfg config.Config

	stop, err := fxapp.PopulateFromApp(ctx, &cfg)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		if resultError != nil {
			stop()
		}
	}()

	db, err := sql.Open("pgx", cfg.DSN)
	if err != nil {
		return nil, nil, err
	}

	return db, stop, nil
}

func mustPrepareGooseAndDB(ctx context.Context) (*sql.DB, func() error) {
	db, stop, err := prepareDB(ctx)
	dieOnError(err)

	err = prepareGoose()
	dieOnError(err)

	return db, stop
}

var migrationUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Database migrations",
	Long:  `Migrate database`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		db, stop := mustPrepareGooseAndDB(ctx)
		defer stop()

		if err := goose.Up(db, "."); err != nil {
			cmd.PrintErr(err)
		}
	},
}

var migrationUpByOneCmd = &cobra.Command{
	Use:   "upbyone",
	Short: "Database migrations",
	Long:  `Migrate database`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		db, stop := mustPrepareGooseAndDB(ctx)
		defer stop()

		if err := goose.UpByOne(db, "."); err != nil {
			cmd.PrintErr(err)
		}
	},
}

var migrationVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Database migrations",
	Long:  `Migrate database`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		db, stop := mustPrepareGooseAndDB(ctx)
		defer stop()

		if err := goose.Version(db, "."); err != nil {
			cmd.PrintErr(err)
		}
	},
}

var migrationStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Database migrations",
	Long:  `Migrate database`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		db, stop := mustPrepareGooseAndDB(ctx)
		defer stop()

		if err := goose.Status(db, "."); err != nil {
			cmd.PrintErr(err)
		}
	},
}

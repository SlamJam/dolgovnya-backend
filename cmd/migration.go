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

	return goose.SetDialect("postgres")
}

func prepareDB(ctx context.Context) (*sql.DB, error) {
	var cfg config.Config
	app := fx.New(
		fxapp.Module,
		fx.Populate(&cfg),
	)

	if err := app.Start(ctx); err != nil {
		return nil, err
	}
	defer app.Stop(ctx)

	db, err := sql.Open("pgx", cfg.DSN)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func dieOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func mustPrepareGooseAndDB(ctx context.Context) *sql.DB {
	db, err := prepareDB(ctx)
	dieOnError(err)

	err = prepareGoose()
	dieOnError(err)
	return db
}

var migrationUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Database migrations",
	Long:  `Migrate database`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var err error
		db := mustPrepareGooseAndDB(ctx)

		err = goose.Up(db, ".")
		dieOnError(err)
	},
}

var migrationUpByOneCmd = &cobra.Command{
	Use:   "upbyone",
	Short: "Database migrations",
	Long:  `Migrate database`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var err error
		db := mustPrepareGooseAndDB(ctx)

		err = goose.UpByOne(db, ".")
		dieOnError(err)
	},
}

var migrationVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Database migrations",
	Long:  `Migrate database`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var err error
		db := mustPrepareGooseAndDB(ctx)

		err = goose.Version(db, ".")
		dieOnError(err)
	},
}

var migrationStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Database migrations",
	Long:  `Migrate database`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var err error
		db := mustPrepareGooseAndDB(ctx)

		err = goose.Status(db, ".")
		dieOnError(err)
	},
}

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
	"go.uber.org/zap"
)

func init() {
	migrationCmd.AddCommand(migrationUpCmd)
	migrationCmd.AddCommand(migrationUpByOneCmd)
	migrationCmd.AddCommand(migrationVersionCmd)
	migrationCmd.AddCommand(migrationStatusCmd)

	rootCmd.AddCommand(migrationCmd)
}

type gooseZap struct {
	zapLogger *zap.SugaredLogger
}

func (gz *gooseZap) Fatal(v ...interface{}) {
	gz.zapLogger.Fatal(v...)
}
func (gz *gooseZap) Fatalf(format string, v ...interface{}) {
	gz.zapLogger.Fatalf(format, v...)
}
func (gz *gooseZap) Print(v ...interface{}) {
	gz.zapLogger.Info(v...)
}
func (gz *gooseZap) Println(v ...interface{}) {
	gz.zapLogger.Info(v...)
}
func (gz *gooseZap) Printf(format string, v ...interface{}) {
	gz.zapLogger.Infof(format, v...)
}

func zapLoggerToGooseLogger(logger *zap.Logger) goose.Logger {
	return &gooseZap{zapLogger: logger.Sugar()}
}

func initGooseAndDB(cfg config.Config, logger *zap.Logger) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.DSN)
	if err != nil {
		return nil, err
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return nil, err
	}

	goose.SetBaseFS(migrations.FS)
	goose.SetLogger(zapLoggerToGooseLogger(logger))

	return db, nil
}

type gooseParams struct {
	fx.In

	Ctx    context.Context
	Cfg    config.Config
	Logger *zap.Logger
}

type gooseCmdFunc func(gooseParams) error

func runGooseCmdInAppContainer(f gooseCmdFunc) (result error) {
	executor := func(p gooseParams) {
		result = f(p)
	}

	fxapp.NewApp(
		fx.Invoke(
			executor,
			func(shutdowner fx.Shutdowner) { shutdowner.Shutdown() },
		),
	).Run()

	return result
}

type gooseSimpleCommand func(*sql.DB, string, ...goose.OptionsFunc) error

func wrapSimpleGooseCommand(f gooseSimpleCommand) gooseCmdFunc {
	return func(p gooseParams) error {
		db, err := initGooseAndDB(p.Cfg, p.Logger)
		if err != nil {
			return err
		}

		return f(db, ".")
	}
}

func execSimpleGooseCommand(f gooseSimpleCommand) error {
	return runGooseCmdInAppContainer(
		wrapSimpleGooseCommand(f),
	)
}

var migrationCmd = &cobra.Command{
	Use:   "migration",
	Short: "Database migrations",
	Long:  `Migrate database`,
}

var migrationUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Migrate the DB to the most recent version available",
	RunE: func(cmd *cobra.Command, args []string) error {
		return execSimpleGooseCommand(goose.Up)
	},
}

var migrationUpByOneCmd = &cobra.Command{
	Use:   "up-by-one",
	Short: "Migrate the DB up by 1",
	RunE: func(cmd *cobra.Command, args []string) error {
		return execSimpleGooseCommand(goose.UpByOne)
	},
}

var migrationStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Dump the migration status for the current DB",
	RunE: func(cmd *cobra.Command, args []string) (resultErr error) {
		return execSimpleGooseCommand(goose.Status)
	},
}

var migrationVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the current version of the database",
	RunE: func(cmd *cobra.Command, args []string) (resultErr error) {
		return execSimpleGooseCommand(goose.Version)
	},
}

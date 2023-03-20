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

func runFuncInAppContainer(f any) {
	fxapp.NewApp(
		fx.Invoke(f, func(shutdowner fx.Shutdowner) { defer shutdowner.Shutdown() }),
	).Run()
}

var migrationCmd = &cobra.Command{
	Use:   "migration",
	Short: "Database migrations",
	Long:  `Migrate database`,
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

var migrationUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Database migrations",
	Long:  `Migrate database`,
	Run: func(cmd *cobra.Command, args []string) {
		runFuncInAppContainer(
			func(ctx context.Context, cfg config.Config, logger *zap.Logger) {
				db, err := initGooseAndDB(cfg, logger)
				if err != nil {
					logger.Fatal("can't init")
				}

				if err := goose.Up(db, "."); err != nil {
					logger.Fatal("fail")
				}
			})
	},
}

var migrationUpByOneCmd = &cobra.Command{
	Use:   "upbyone",
	Short: "Database migrations",
	Long:  `Migrate database`,
	Run: func(cmd *cobra.Command, args []string) {
		runFuncInAppContainer(
			func(ctx context.Context, cfg config.Config, logger *zap.Logger) {
				db, err := initGooseAndDB(cfg, logger)
				if err != nil {
					logger.Fatal("can't init")
				}

				if err := goose.UpByOne(db, "."); err != nil {
					logger.Fatal("fail")
				}
			})
	},
}

var migrationVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Database migrations",
	Long:  `Migrate database`,
	Run: func(cmd *cobra.Command, args []string) {
		runFuncInAppContainer(
			func(ctx context.Context, cfg config.Config, logger *zap.Logger) {
				db, err := initGooseAndDB(cfg, logger)
				if err != nil {
					logger.Fatal("can't init")
				}

				if err := goose.Version(db, "."); err != nil {
					logger.Fatal("fail")
				}
			})
	},
}

var migrationStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Database migrations",
	Long:  `Migrate database`,
	Run: func(cmd *cobra.Command, args []string) {
		runFuncInAppContainer(
			func(ctx context.Context, cfg config.Config, logger *zap.Logger) {
				db, err := initGooseAndDB(cfg, logger)
				if err != nil {
					logger.Fatal("can't init")
				}

				if err := goose.Status(db, "."); err != nil {
					logger.Fatal("fail")
				}
			})
	},
}

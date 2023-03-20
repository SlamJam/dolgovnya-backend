package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/SlamJam/dolgovnya-backend/internal/app/config"
	"github.com/SlamJam/dolgovnya-backend/internal/bootstrap/fxapp"
	"github.com/SlamJam/dolgovnya-backend/migrations"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog"
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

type gooseLog struct {
	// zapLogger *zap.SugaredLogger
	logger zerolog.Logger
}

// getMessage format with Sprint, Sprintf, or neither.
func getMessage(template string, fmtArgs []interface{}) string {
	if len(fmtArgs) == 0 {
		return template
	}

	if template != "" {
		return fmt.Sprintf(template, fmtArgs...)
	}

	if len(fmtArgs) == 1 {
		if str, ok := fmtArgs[0].(string); ok {
			return str
		}
	}
	return fmt.Sprint(fmtArgs...)
}

// getMessageln format with Sprintln.
func getMessageln(fmtArgs []interface{}) string {
	msg := fmt.Sprintln(fmtArgs...)
	return msg[:len(msg)-1]
}

func (l *gooseLog) Fatal(v ...interface{}) {
	l.logger.Fatal().Msg(getMessage("", v))
}
func (gz *gooseLog) Fatalf(format string, v ...interface{}) {
	gz.logger.Fatal().Msg(getMessage(format, v))
}
func (gz *gooseLog) Print(v ...interface{}) {
	gz.logger.Info().Msg(getMessage("", v))
}
func (gz *gooseLog) Println(v ...interface{}) {
	gz.logger.Info().Msg(getMessageln(v))
}
func (gz *gooseLog) Printf(format string, v ...interface{}) {
	gz.logger.Info().Msg(getMessage(strings.TrimSpace(format), v))
}

func newGooseLogger(logger zerolog.Logger) goose.Logger {
	return &gooseLog{logger: logger}
}

func runCmdInAppContainer[T any](cmd func(T) error) (result error) {
	fxapp.NewApp(
		fx.Provide(newGooseLogger),
		fx.Invoke(
			func(params T) {
				result = cmd(params)
			},
			func(shutdowner fx.Shutdowner) {
				shutdowner.Shutdown()
			},
		),
	).Run()

	return result
}

type gooseParams struct {
	fx.In

	Ctx    context.Context
	Cfg    config.Config
	Logger goose.Logger
}

type gooseSimpleCommand func(*sql.DB, string, ...goose.OptionsFunc) error

func execSimpleGooseCommand(gooseCmd gooseSimpleCommand) error {
	return runCmdInAppContainer(
		func(p gooseParams) error {
			// goose.OpenDBWithDriver()
			db, err := sql.Open("pgx", p.Cfg.DSN)
			if err != nil {
				return err
			}

			if err := goose.SetDialect("postgres"); err != nil {
				return err
			}

			goose.SetBaseFS(migrations.FS)
			goose.SetLogger(p.Logger)
			goose.SetVerbose(verbose)

			return gooseCmd(db, ".")
		},
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

package cmd

import (
	"database/sql"

	"github.com/SlamJam/dolgovnya-backend/migrations"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(migrationCmd)
}

var migrationCmd = &cobra.Command{
	Use:   "migration",
	Short: "Database migrations",
	Long:  `Migrate database`,
	Run: func(cmd *cobra.Command, args []string) {
		goose.SetBaseFS(migrations.FS)

		if err := goose.SetDialect("postgres"); err != nil {
			panic(err)
		}

		// setup database
		var db *sql.DB

		if err := goose.Up(db, "."); err != nil {
			panic(err)
		}
	},
}

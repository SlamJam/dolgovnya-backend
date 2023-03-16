package pgsql

import (
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

var (
	psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
)

type Storage struct {
	pool *sqlx.DB
}

func NewStorage(dsn string) (*Storage, error) {
	db, err := GetDB(dsn)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &Storage{
		pool: db,
	}, nil
}

func GetDB(uri string) (*sqlx.DB, error) {
	// before : directly using sqlx
	// DB, err = sqlx.Connect("postgres", uri)
	// after : using pgx to setup connection
	DB, err := PgxCreateDB(uri)
	if err != nil {
		return nil, err
	}
	DB.SetMaxIdleConns(4)
	DB.SetMaxOpenConns(4)
	DB.SetConnMaxLifetime(time.Duration(30) * time.Minute)

	return DB, nil
}

func PgxCreateDB(uri string) (*sqlx.DB, error) {
	connConfig, err := pgx.ParseConfig(uri)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// afterConnect := stdlib.OptionAfterConnect(func(ctx context.Context, conn *pgx.Conn) error {
	// 	_, err := conn.Exec(ctx, `
	// 		 SET SESSION "some.key" = 'somekey';
	// 		 CREATE TEMP TABLE IF NOT EXISTS sometable AS SELECT 212 id;
	// 	`)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	return nil
	// })

	pgxdb := stdlib.OpenDB(*connConfig)
	return sqlx.NewDb(pgxdb, "pgx"), nil
}

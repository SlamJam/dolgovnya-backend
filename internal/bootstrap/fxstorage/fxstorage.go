package fxstorage

import (
	"github.com/SlamJam/dolgovnya-backend/internal/app/storage/pgsql"
	"go.uber.org/fx"
)

func NewPgStorage(lc fx.Lifecycle) (*pgsql.Storage, error) {
	dsn := "postgres://"

	s, err := pgsql.NewStorage(dsn)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: s.Close,
	})

	return s, nil
}

var Module = fx.Module("pgsql",
	fx.Provide(NewPgStorage),
)

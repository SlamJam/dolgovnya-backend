package fxstorage

import (
	"github.com/SlamJam/dolgovnya-backend/internal/app/config"
	"github.com/SlamJam/dolgovnya-backend/internal/app/services"
	"github.com/SlamJam/dolgovnya-backend/internal/app/storage/pgsql"
	"go.uber.org/fx"
)

func NewPgStorage(lc fx.Lifecycle, cfg config.Config) (*pgsql.Storage, error) {
	s, err := pgsql.NewStorage(cfg.DSN)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: s.Close,
	})

	return s, nil
}

func newSplitTheBillStorage(lc fx.Lifecycle, s *pgsql.Storage) services.SplitTheBillStorage {
	return s
}

var Module = fx.Module("pgsql",
	fx.Provide(NewPgStorage),
	fx.Provide(newSplitTheBillStorage),
)

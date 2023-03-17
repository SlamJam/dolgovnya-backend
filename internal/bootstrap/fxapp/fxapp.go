package fxapp

import (
	"github.com/SlamJam/dolgovnya-backend/internal/app/services"
	"github.com/SlamJam/dolgovnya-backend/internal/app/storage/pgsql"
	"github.com/SlamJam/dolgovnya-backend/internal/bootstrap/fxstorage"
	"go.uber.org/fx"
)

func NewSplitTheBillStorage(lc fx.Lifecycle, s *pgsql.Storage) services.SplitTheBillStorage {
	return s
}

var Module = fx.Module("app",
	fxstorage.Module,
	fx.Provide(NewSplitTheBillStorage),
)

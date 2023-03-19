package fxapp

import (
	"context"

	"github.com/SlamJam/dolgovnya-backend/internal/app/services"
	"github.com/SlamJam/dolgovnya-backend/internal/app/storage/pgsql"
	"github.com/SlamJam/dolgovnya-backend/internal/bootstrap/fxconfig"
	"github.com/SlamJam/dolgovnya-backend/internal/bootstrap/fxstorage"
	"go.uber.org/fx"
)

func newSplitTheBillStorage(lc fx.Lifecycle, s *pgsql.Storage) services.SplitTheBillStorage {
	return s
}

var Module = fx.Module("app",
	fxstorage.Module,
	fxconfig.Module,
	fx.Provide(newSplitTheBillStorage),
)

func PopulateFromApp(ctx context.Context, pointers ...any) (func() error, error) {
	opts := make([]fx.Option, 0, len(pointers)+1)
	opts = append(opts, Module)

	for _, p := range pointers {
		opts = append(opts, fx.Populate(p))
	}

	app := fx.New(opts...)

	if err := app.Start(ctx); err != nil {
		return nil, err
	}

	return func() error {
		return app.Stop(ctx)
	}, nil
}

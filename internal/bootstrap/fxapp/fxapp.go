package fxapp

import (
	"context"

	"github.com/SlamJam/dolgovnya-backend/internal/bootstrap/fxconfig"
	"github.com/SlamJam/dolgovnya-backend/internal/bootstrap/fxstorage"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewApp(opts ...fx.Option) *fx.App {
	return fx.New(
		append(opts,
			Module,
			// Заглушаем всё, что ниже WARN
			fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
				return &fxevent.ZapLogger{Logger: log.WithOptions(zap.IncreaseLevel(zap.WarnLevel))}
			}),
		)...,
	)
}

var Module = fx.Module("app",
	fxstorage.Module,
	fxconfig.Module,
	fx.Provide(NewZapLogger),
	fx.Provide(NewContext),
)

func NewContext(lc fx.Lifecycle) context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			cancel()
			return nil
		},
	})

	return ctx
}

func NewZapLogger(lc fx.Lifecycle) (*zap.Logger, error) {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			_ = logger.Sync()
			return nil
		},
	})

	return logger, nil
}

func PopulateFromApp(ctx context.Context, pointers ...any) (func() error, error) {
	opts := make([]fx.Option, 0, len(pointers)+1)
	opts = append(opts, Module)

	for _, p := range pointers {
		opts = append(opts, fx.Populate(p))
	}

	app := NewApp(opts...)

	if err := app.Start(ctx); err != nil {
		return nil, err
	}

	return func() error {
		return app.Stop(ctx)
	}, nil
}

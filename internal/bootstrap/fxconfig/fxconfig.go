package fxconfig

import (
	"github.com/SlamJam/dolgovnya-backend/internal/app/config"
	"go.uber.org/fx"
)

func NewConfig() (config.Config, error) {
	return config.Config{
		DSN: "postgresql://postgres@localhost",
	}, nil
}

var Module = fx.Module("config",
	fx.Provide(NewConfig),
)

package fxcli

import (
	"time"

	"github.com/SlamJam/dolgovnya-backend/cmd/cli"
	"github.com/SlamJam/dolgovnya-backend/internal/app/config"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
)

var Module = fx.Module("cli",
	fx.Provide(NewCliLogger),
)

func NewCliLogger(cfg config.Config) cli.Logger {
	var logger zerolog.Logger

	output := zerolog.NewConsoleWriter()
	output.PartsOrder = []string{
		zerolog.MessageFieldName,
	}
	output.TimeFormat = time.TimeOnly
	logger = zerolog.New(output)

	return cli.Logger{Logger: logger}
}

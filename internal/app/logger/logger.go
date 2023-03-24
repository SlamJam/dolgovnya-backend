package logger

import (
	"context"

	"github.com/rs/zerolog"
)

type Logger = *zerolog.Logger

func FromCtx(ctx context.Context) Logger {
	return nil
}

func FromCtxOrDefault(ctx context.Context, defaultLogger Logger) Logger {
	if log := FromCtx(ctx); log != nil {
		return log
	}

	return defaultLogger
}

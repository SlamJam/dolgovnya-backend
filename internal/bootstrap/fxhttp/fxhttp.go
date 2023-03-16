package fxhttp

import (
	"go.uber.org/fx"
)

var Module = fx.Module("http",
	fx.Provide(NewHTTPServer),
	fx.Provide(NewConnectServer),
)

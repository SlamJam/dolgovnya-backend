package fxhttp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/SlamJam/dolgovnya-backend/internal/components"
	connect_handlers "github.com/SlamJam/dolgovnya-backend/internal/connect-handlers"
	"github.com/SlamJam/dolgovnya-backend/internal/pb/pbconnect"
	"github.com/SlamJam/go-libs/component"
	"go.uber.org/fx"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type ConnectServer component.Component

func NewConnectServer(lc fx.Lifecycle) ConnectServer {
	addr := ":8085"
	mux := http.NewServeMux()
	// The generated constructors return a path and a plain net/http handler.
	mux.Handle(pbconnect.NewSplitTheBillServiceHandler(&connect_handlers.SplitTheBillServiceHandler{}))

	// For gRPC clients, it's convenient to support HTTP/2 without TLS. You can
	// avoid x/net/http2 by using http.ListenAndServeTLS.
	c := components.NewHttpServer(addr, h2c.NewHandler(mux, &http2.Server{}))

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			fmt.Println("Starting Connect server at", addr)
			c.Start(ctx)
			return nil
		},
		OnStop: c.Interrupt,
	})

	return c
}

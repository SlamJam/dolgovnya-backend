package fxhttp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/SlamJam/dolgovnya-backend/internal/components"
	"github.com/SlamJam/go-libs/component"
	"go.uber.org/fx"
)

const (
	swaggerDefinitionsDir = "definitions"
)

type HTTPServer component.Component

func NewHTTPServer(lc fx.Lifecycle) HTTPServer {
	addr := ":8080"
	mux := http.NewServeMux()
	mux.Handle("/", NewSwaggerUIHandler(swaggerDefinitionsDir))

	c := components.NewHttpServer(addr, mux)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			c.Start(ctx)
			fmt.Println("Starting HTTP server at", addr)
			return nil
		},
		OnStop: c.Interrupt,
	})

	return c
}

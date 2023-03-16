package components

import (
	"context"
	"net/http"

	"github.com/SlamJam/go-libs/component"
)

type httpServer struct {
	component.Component
}

func NewHttpServer(addr string, handler http.Handler) component.Component {
	c := &httpServer{}

	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	c.Component = component.NewComponent(
		func(ctx context.Context) error {
			return server.ListenAndServe()
		},
		component.WithOnInterrupt(func(ctx context.Context) error {
			return server.Shutdown(ctx)
		}),
	)

	return c
}

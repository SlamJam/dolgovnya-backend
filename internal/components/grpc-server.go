package components

import (
	"github.com/SlamJam/go-libs/component"
)

type grpcServer struct {
	component.Component
}

func NewGrpcServer() component.Component {
	c := &grpcServer{}
	// c.Component = component.NewComponent()

	return c
}

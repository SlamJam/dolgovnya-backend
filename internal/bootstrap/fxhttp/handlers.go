package fxhttp

import (
	"net/http"

	"github.com/SlamJam/dolgovnya-backend/internal/swagger"
	"github.com/swaggest/swgui"
	"github.com/swaggest/swgui/v4emb"
)

func NewSwaggerUIHandler(definitionsDir string) http.Handler {
	config := swgui.Config{
		Title: "Dolgovnya's API",
		SettingsUI: map[string]string{
			"spec": swagger.SwaggerJson,
		},
	}

	return v4emb.NewHandlerWithConfig(config)
}

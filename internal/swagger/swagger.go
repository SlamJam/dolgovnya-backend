package swagger

import (
	_ "embed"
)

//go:embed apidocs.swagger.json
var SwaggerJson string

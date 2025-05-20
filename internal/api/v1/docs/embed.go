package docs

import "embed"

//go:embed swagger-ui/* openapi.yaml
var ApiFiles embed.FS

// Package api содержит сгенерированный HTTP-контракт TourneyHub v1.
// Изменения вносятся в openapi.yaml; команда make generate обновляет api.gen.go.
package api

import _ "embed"

// Specification — исходный OpenAPI YAML, который публикуется через HTTP.
//
//go:embed openapi.yaml
var Specification []byte

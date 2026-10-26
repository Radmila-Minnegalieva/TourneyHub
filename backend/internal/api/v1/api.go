// Package api реализует транспортный слой HTTP API TourneyHub.
package api

import (
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/config"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/service"
	contract "github.com/Radmila-Minnegalieva/TourneyHub/backend/pkg/api/v1"
)

// API преобразует HTTP DTO и вызывает сценарии сервисного слоя.
// Объекты PostgreSQL и SQL-запросы транспортному слою не передаются.
type API struct {
	identity    service.Identity
	teams       service.Teams
	tournaments service.Tournaments
	competition service.Competition
	admin       service.Admin
	catalog     service.Catalog
	resources   service.Resources
	available   bool
	cfg         config.HTTP
}

var _ contract.ServerInterface = (*API)(nil)

func New(backend service.Backend, cfg ...config.HTTP) *API {
	httpConfig := config.Default().HTTP
	if len(cfg) > 0 {
		httpConfig = cfg[0]
	}
	return &API{
		identity:    backend,
		teams:       backend,
		tournaments: backend,
		competition: backend,
		admin:       backend,
		catalog:     backend,
		resources:   backend,
		available:   backend != nil,
		cfg:         httpConfig,
	}
}

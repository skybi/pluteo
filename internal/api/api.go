package api

import (
	"errors"
	"github.com/skybi/data-server/internal/api/portal"
	"github.com/skybi/data-server/internal/config"
	"github.com/skybi/data-server/internal/storage"
	"net/http"
)

// Service represents the portal & data API service
type Service struct {
	Config  *config.Config
	Storage storage.Driver
	portal  *portal.Service
}

// Startup starts up the portal & data APIs
func (service *Service) Startup(errs chan<- error) {
	portalService := &portal.Service{
		Config:  service.Config,
		Storage: service.Storage,
	}
	service.portal = portalService
	go func() {
		if err := portalService.Startup(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errs <- err
		}
	}()
}

// Shutdown shuts down the portal & data APIs
func (service *Service) Shutdown() {
	if service.portal != nil {
		service.portal.Shutdown()
		service.portal = nil
	}
}

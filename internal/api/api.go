package api

import (
	"errors"
	"github.com/skybi/pluteo/internal/api/data"
	"github.com/skybi/pluteo/internal/api/portal"
	"github.com/skybi/pluteo/internal/apikey/quota"
	"github.com/skybi/pluteo/internal/config"
	"github.com/skybi/pluteo/internal/storage"
	"net/http"
)

// Service represents the portal & data API service
type Service struct {
	Config       *config.Config
	Storage      storage.Driver
	QuotaTracker *quota.Tracker

	portal *portal.Service
	data   *data.Service
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

	dataService := &data.Service{
		Config:       service.Config,
		Storage:      service.Storage,
		QuotaTracker: service.QuotaTracker,
	}
	service.data = dataService
	go func() {
		if err := dataService.Startup(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
	if service.data != nil {
		service.data.Shutdown()
		service.data = nil
	}
}

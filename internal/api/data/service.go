package data

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog/log"
	"github.com/skybi/data-server/internal/api/schema"
	"github.com/skybi/data-server/internal/apikey"
	"github.com/skybi/data-server/internal/apikey/quota"
	"github.com/skybi/data-server/internal/config"
	"github.com/skybi/data-server/internal/function"
	"github.com/skybi/data-server/internal/storage"
	"net/http"
)

// Service represents the data API service
type Service struct {
	server *http.Server

	Config       *config.Config
	Storage      storage.Driver
	QuotaTracker *quota.Tracker

	writer *schema.Writer
}

// Startup starts up the data API
func (service *Service) Startup() error {
	// Create the HTTP schema writer
	service.writer = &schema.Writer{
		InternalErrorHook: func(err error) {
			log.Error().Err(err).Msg("the data API experienced an unexpected error")
		},
	}

	// Create the HTTP router
	router := chi.NewRouter()
	router.Use(middleware.RedirectSlashes)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://*", "https://*"},
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	}))
	router.NotFound(func(writer http.ResponseWriter, _ *http.Request) {
		service.writer.WriteErrors(writer, http.StatusNotFound, schema.ErrNotFound)
	})
	router.MethodNotAllowed(func(writer http.ResponseWriter, _ *http.Request) {
		service.writer.WriteErrors(writer, http.StatusMethodNotAllowed, schema.ErrMethodNotAllowed)
	})

	// Register the API endpoint handlers
	service.registerEndpoints(router)

	// Start up the server
	server := &http.Server{
		Addr:    service.Config.DataAPIListenAddress,
		Handler: router,
	}
	service.server = server
	return server.ListenAndServe()
}

// Shutdown shuts down the portal API
func (service *Service) Shutdown() {
	if service.server != nil {
		service.server.Close()
		service.server = nil
	}
}

func (service *Service) registerEndpoints(router chi.Router) {
	// Register the key information endpoint
	router.Get("/v1/key_info", function.Nest[http.HandlerFunc](
		service.EndpointGetKeyInfo,
		service.MiddlewareVerifyKey,
	))

	// Register the METAR controller endpoints
	router.Get("/v1/metars", function.Nest[http.HandlerFunc](
		service.EndpointGetMETARs,
		service.MiddlewareVerifyKey,
		service.MiddlewareVerifyKeyCapabilities(apikey.CapabilityReadMETARs),
		service.MiddlewareVerifyKeyQuota,
	))
	router.Get("/v1/metars/{id}", function.Nest[http.HandlerFunc](
		service.EndpointGetMETAR,
		service.MiddlewareVerifyKey,
		service.MiddlewareVerifyKeyCapabilities(apikey.CapabilityReadMETARs),
		service.MiddlewareVerifyKeyQuota,
	))
}

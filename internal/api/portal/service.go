package portal

import (
	"github.com/go-chi/chi/v5"
	"github.com/skybi/data-server/internal/config"
	"net/http"
)

// Service represents the portal API service
type Service struct {
	server *http.Server
	Config *config.Config
}

// Startup starts up the portal API
func (service *Service) Startup() error {
	router := chi.NewRouter()

	// TODO: Register actual handlers
	router.Get("/*", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("Hello, world!"))
	})

	server := &http.Server{
		Addr:    service.Config.PortalAPIAddress,
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

package portal

import (
	"context"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/skybi/data-server/internal/api/portal/session"
	"github.com/skybi/data-server/internal/api/portal/session/storage/inmem"
	"github.com/skybi/data-server/internal/api/schema"
	"github.com/skybi/data-server/internal/config"
	"github.com/skybi/data-server/internal/storage"
	"golang.org/x/oauth2"
	"net/http"
)

// Service represents the portal API service
type Service struct {
	server *http.Server

	Config *config.Config

	Storage storage.Driver

	oidcOAuth2Config    *oauth2.Config
	oidcProvider        *oidc.Provider
	oidcIDTokenVerifier *oidc.IDTokenVerifier
	sessionStorage      session.Storage

	writer *schema.Writer
}

// Startup starts up the portal API
func (service *Service) Startup() error {
	// Create the HTTP schema writer
	service.writer = &schema.Writer{
		InternalErrorHook: func(err error) {
			log.Error().Err(err).Msg("the portal API experienced an unexpected error")
		},
	}

	// Create the HTTP router
	router := chi.NewRouter()
	router.NotFound(func(writer http.ResponseWriter, _ *http.Request) {
		service.writer.WriteErrors(writer, http.StatusNotFound, schema.ErrNotFound)
	})
	router.MethodNotAllowed(func(writer http.ResponseWriter, _ *http.Request) {
		service.writer.WriteErrors(writer, http.StatusMethodNotAllowed, schema.ErrMethodNotAllowed)
	})

	// Create the OIDC provider & ID token verifier
	oidcProvider, err := oidc.NewProvider(context.Background(), service.Config.OIDCProviderURL)
	if err != nil {
		return err
	}
	service.oidcProvider = oidcProvider
	service.oidcIDTokenVerifier = oidcProvider.Verifier(&oidc.Config{
		ClientID: service.Config.OIDCClientID,
	})

	// Create the OAuth2 config
	service.oidcOAuth2Config = &oauth2.Config{
		ClientID:     service.Config.OIDCClientID,
		ClientSecret: service.Config.OIDCClientSecret,
		Endpoint:     oidcProvider.Endpoint(),
		RedirectURL:  service.Config.PortalAPIBaseAddress + "/v1/auth/oidc/callback",
		Scopes:       []string{oidc.ScopeOpenID},
	}

	// Create the session storage
	sessionStorage, err := inmem.New()
	if err != nil {
		return err
	}
	service.sessionStorage = sessionStorage

	// Register the OIDC authentication endpoints
	router.Get("/v1/auth/oidc/login_flow", service.EndpointOIDCLoginFlow)
	router.Get("/v1/auth/oidc/callback", service.EndpointOIDCLoginCallback)
	// TODO: Implement backchannel logout

	// Start up the server
	server := &http.Server{
		Addr:    service.Config.PortalAPIListenAddress,
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

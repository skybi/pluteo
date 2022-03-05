package portal

import (
	"context"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-chi/chi/v5"
	"github.com/skybi/data-server/internal/config"
	"golang.org/x/oauth2"
	"net/http"
)

// Service represents the portal API service
type Service struct {
	server *http.Server

	Config *config.Config

	oidcOAuth2Config    *oauth2.Config
	oidcIDTokenVerifier *oidc.IDTokenVerifier
}

// Startup starts up the portal API
func (service *Service) Startup() error {
	router := chi.NewRouter()
	router.NotFound(func(writer http.ResponseWriter, _ *http.Request) {
		service.error(writer, http.StatusNotFound, "not found")
	})
	router.MethodNotAllowed(func(writer http.ResponseWriter, _ *http.Request) {
		service.error(writer, http.StatusMethodNotAllowed, "method not allowed")
	})

	// Create the OIDC provider & ID token verifier
	oidcProvider, err := oidc.NewProvider(context.Background(), service.Config.OIDCProviderURL)
	if err != nil {
		return err
	}
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

	// Register the OIDC authentication endpoints
	router.Get("/v1/auth/oidc/login_flow", service.EndpointOIDCLoginFlow)
	router.Get("/v1/auth/oidc/callback", service.EndpointOIDCLoginCallback)
	router.Post("/v1/auth/oidc/backchannel_logout", service.EndpointOIDCBackchannelLogout)

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

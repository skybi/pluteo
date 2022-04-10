package portal

import (
	"context"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog/log"
	"github.com/skybi/data-server/internal/api/portal/session"
	"github.com/skybi/data-server/internal/api/portal/session/storage/inmem"
	"github.com/skybi/data-server/internal/api/schema"
	"github.com/skybi/data-server/internal/config"
	"github.com/skybi/data-server/internal/function"
	"github.com/skybi/data-server/internal/storage"
	"golang.org/x/oauth2"
	"net/http"
)

// Service represents the portal API service
type Service struct {
	server *http.Server

	Config  *config.Config
	Storage storage.Driver

	oidcOAuth2Config        *oauth2.Config
	oidcProvider            *oidc.Provider
	oidcIDTokenVerifier     *oidc.IDTokenVerifier
	oidcLogoutTokenVerifier *oidc.LogoutTokenVerifier
	sessionStorage          session.Storage

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
	router.Use(middleware.RedirectSlashes)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{service.Config.PortalAPIAllowedOrigin},
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

	// Create the OIDC provider, ID token verifier & logout token verifier
	oidcProvider, err := oidc.NewProvider(context.Background(), service.Config.OIDCProviderURL)
	if err != nil {
		return err
	}
	service.oidcProvider = oidcProvider
	cfg := &oidc.Config{
		ClientID: service.Config.OIDCClientID,
	}
	service.oidcIDTokenVerifier = oidcProvider.Verifier(cfg)
	service.oidcLogoutTokenVerifier = oidcProvider.LogoutVerifier(cfg)

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

	// Register the API endpoint handlers
	service.registerEndpoints(router)

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

func (service *Service) registerEndpoints(router chi.Router) {
	// Register the OIDC authentication endpoints
	router.Get("/v1/auth/oidc/login_flow", service.EndpointOIDCLoginFlow)
	router.Get("/v1/auth/oidc/callback", service.EndpointOIDCLoginCallback)
	router.Post("/v1/auth/oidc/backchannel_logout", service.EndpointOIDCBackchannelLogout)

	// Register the user controller endpoints
	router.Get("/v1/users", function.Nest[http.HandlerFunc](
		service.EndpointGetUsers,
		service.MiddlewareVerifySession,
		service.MiddlewareFetchUser,
		service.MiddlewareCheckAdmin,
	))
	router.Get("/v1/users/{id}", function.Nest[http.HandlerFunc](
		service.EndpointGetUser,
		service.MiddlewareVerifySession,
		service.MiddlewareFetchUser,
		service.MiddlewareCheckAdmin,
	))
	router.Patch("/v1/users/{id}", function.Nest[http.HandlerFunc](
		service.EndpointEditUser,
		service.MiddlewareVerifySession,
		service.MiddlewareFetchUser,
		service.MiddlewareCheckAdmin,
	))
	router.Delete("/v1/users/{id}", function.Nest[http.HandlerFunc](
		service.EndpointDeleteUserData,
		service.MiddlewareVerifySession,
		service.MiddlewareFetchUser,
		service.MiddlewareCheckAdmin,
	))
	router.Get("/v1/me", function.Nest[http.HandlerFunc](
		service.EndpointGetSelfUser,
		service.MiddlewareVerifySession,
		service.MiddlewareFetchUser,
	))
	router.Delete("/v1/me", function.Nest[http.HandlerFunc](
		service.EndpointDeleteSelfUserData,
		service.MiddlewareVerifySession,
		service.MiddlewareFetchUser,
	))

	// Register the API key controller endpoints
	router.Post("/v1/api_keys", function.Nest[http.HandlerFunc](
		service.EndpointCreateAPIKey,
		service.MiddlewareVerifySession,
		service.MiddlewareFetchUser,
	))
	router.Get("/v1/api_keys", function.Nest[http.HandlerFunc](
		service.EndpointGetAPIKeys,
		service.MiddlewareVerifySession,
		service.MiddlewareFetchUser,
	))
	router.Get("/v1/api_keys/{id}", function.Nest[http.HandlerFunc](
		service.EndpointGetAPIKey,
		service.MiddlewareVerifySession,
		service.MiddlewareFetchUser,
	))
	router.Patch("/v1/api_keys/{id}", function.Nest[http.HandlerFunc](
		service.EndpointEditAPIKey,
		service.MiddlewareVerifySession,
		service.MiddlewareFetchUser,
	))
	router.Delete("/v1/api_keys/{id}", function.Nest[http.HandlerFunc](
		service.EndpointDeleteAPIKey,
		service.MiddlewareVerifySession,
		service.MiddlewareFetchUser,
	))
}

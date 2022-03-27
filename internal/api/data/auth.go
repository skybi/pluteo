package data

import (
	"context"
	"errors"
	"github.com/skybi/data-server/internal/api/schema"
	"github.com/skybi/data-server/internal/apikey"
	"github.com/skybi/data-server/internal/bitflag"
	"net/http"
	"strings"
)

var contextValueKey = "key"

var errAuthInsufficientCapabilities = func(provided, required bitflag.Container) *schema.Error {
	return &schema.Error{
		Type:    "data.access.insufficientKeyCapabilities",
		Message: "The specified API key lacks at least one capability required for this action.",
		Details: map[string]any{
			"provided": provided,
			"required": required,
			"missing":  required & ^provided,
		},
	}
}

// MiddlewareVerifyKey makes sure that the requesting client has provided a valid API key.
// Additionally, it injects the API key object itself into the request context.
func (service *Service) MiddlewareVerifyKey(next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		// Try to read the 'Authorization' header and verify it is of type 'Bearer'
		header := request.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer") {
			service.writer.WriteErrors(writer, http.StatusUnauthorized, schema.ErrUnauthorized)
			return
		}

		// Try to retrieve the API key out of the database
		rawKey := strings.TrimSpace(strings.TrimPrefix(header, "Bearer"))
		key, err := service.Storage.APIKeys().GetByRawKey(request.Context(), rawKey)
		if err != nil {
			service.writer.WriteInternalError(writer, err)
			return
		}
		if key == nil {
			service.writer.WriteErrors(writer, http.StatusUnauthorized, schema.ErrUnauthorized)
			return
		}

		// Delegate to the next handler
		request = request.WithContext(context.WithValue(request.Context(), contextValueKey, key))
		next(writer, request)
	}
}

// MiddlewareVerifyKeyCapabilities makes sure that the provided API key has a set of required capabilities
func (service *Service) MiddlewareVerifyKeyCapabilities(caps ...bitflag.Flag) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(writer http.ResponseWriter, request *http.Request) {
			// Extract the API key object
			key, ok := request.Context().Value(contextValueKey).(*apikey.Key)
			if !ok {
				service.writer.WriteInternalError(writer, errors.New("API key capability check without API key verification"))
				return
			}

			// Verify the keys capabilities
			if !key.Capabilities.Has(caps...) {
				err := errAuthInsufficientCapabilities(key.Capabilities, bitflag.EmptyContainer.With(caps...))
				service.writer.WriteErrors(writer, http.StatusForbidden, err)
				return
			}
			// Delegate to the next handler
			next(writer, request)
		}
	}
}

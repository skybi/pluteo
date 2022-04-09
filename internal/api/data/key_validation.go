package data

import (
	"context"
	"errors"
	"fmt"
	"github.com/skybi/data-server/internal/api/schema"
	"github.com/skybi/data-server/internal/apikey"
	"github.com/skybi/data-server/internal/bitflag"
	"net/http"
	"strings"
)

var contextValueKey = "key"

var (
	errKeyInsufficientCapabilities = func(provided, required bitflag.Container) *schema.Error {
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
	errKeyNoQuotaLeft = &schema.Error{
		Type:    "data.access.noKeyQuotaLeft",
		Message: "The specified API key has no API quota left.",
		Details: nil,
	}
	errKeyRateLimitExceeded = func(max int) *schema.Error {
		return &schema.Error{
			Type:    "data.access.rateLimitExceeded",
			Message: fmt.Sprintf("The specified API key is being rate limited (max. %d requests per minute).", max),
			Details: map[string]any{
				"max": max,
			},
		}
	}
)

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

			// Verify the key's capabilities
			if !key.Capabilities.Has(caps...) {
				err := errKeyInsufficientCapabilities(key.Capabilities, bitflag.EmptyContainer.With(caps...))
				service.writer.WriteErrors(writer, http.StatusForbidden, err)
				return
			}

			// Delegate to the next handler
			next(writer, request)
		}
	}
}

// MiddlewareVerifyKeyQuota makes sure that the provided API key has quota left to perform a request
func (service *Service) MiddlewareVerifyKeyQuota(next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		// Extract the API key object
		key, ok := request.Context().Value(contextValueKey).(*apikey.Key)
		if !ok {
			service.writer.WriteInternalError(writer, errors.New("API key quota check without API key verification"))
			return
		}

		// Verify the key's quota
		if key.Quota >= 0 && service.QuotaTracker.Get(key) >= key.Quota {
			service.writer.WriteErrors(writer, http.StatusTooManyRequests, errKeyNoQuotaLeft)
			return
		}

		// Delegate to the next handler
		next(writer, request)
	}
}

// MiddlewareVerifyKeyRateLimit makes sure that the provided API key does not make too many requests per minute
func (service *Service) MiddlewareVerifyKeyRateLimit(next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		// Extract the API key object
		key, ok := request.Context().Value(contextValueKey).(*apikey.Key)
		if !ok {
			service.writer.WriteInternalError(writer, errors.New("API key rate limit check without API key verification"))
			return
		}

		// Make sure the client may make another request
		used, ok := service.requestCounter.Lookup(key.ID)
		if !ok {
			used = 0
		}
		if key.RateLimit >= 0 && int(used) >= key.RateLimit {
			service.writer.WriteErrors(writer, http.StatusTooManyRequests, errKeyRateLimitExceeded(key.RateLimit))
			return
		}

		// Update the client's request counter
		service.requestCounter.Set(key.ID, used+1)

		// Delegate to the next handler
		next(writer, request)
	}
}

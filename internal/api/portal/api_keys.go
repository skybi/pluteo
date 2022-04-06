package portal

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/skybi/data-server/internal/api/schema"
	"github.com/skybi/data-server/internal/apikey"
	"github.com/skybi/data-server/internal/bitflag"
	"github.com/skybi/data-server/internal/user"
	"math"
	"net/http"
)

var (
	errAPIKeyQuotaNotAllowed = func(requested, max int64) *schema.Error {
		return &schema.Error{
			Type:    "portal.apiKey.quotaNotAllowed",
			Message: fmt.Sprintf("The requested API key quota (%d) is not allowed by the clients API key policy (max allowed: %d).", requested, max),
			Details: map[string]any{
				"requested":   requested,
				"max_allowed": max,
			},
		}
	}
	errAPIKeyRateLimitNotAllowed = func(requested, max int) *schema.Error {
		return &schema.Error{
			Type:    "portal.apiKey.rateLimitNotAllowed",
			Message: fmt.Sprintf("The requested API key rate limit (%d) is not allowed by the clients API key policy (max allowed: %d).", requested, max),
			Details: map[string]any{
				"requested":   requested,
				"max_allowed": max,
			},
		}
	}
	errAPIKeyCapabilitiesNotAllowed = func(requested, allowed bitflag.Container) *schema.Error {
		return &schema.Error{
			Type:    "portal.apiKey.capabilitiesNotAllowed",
			Message: "The requested API key capabilities are not allowed by the clients API key policy.",
			Details: map[string]any{
				"requested":  requested,
				"allowed":    allowed,
				"disallowed": requested & ^allowed,
			},
		}
	}
)

type endpointCreateAPIKeyRequestPayload struct {
	Description  *string            `json:"description"`
	Quota        *int64             `json:"quota" required:"true"`
	RateLimit    *int               `json:"rate_limit" required:"true"`
	Capabilities *bitflag.Container `json:"capabilities" required:"true"`
}

type endpointCreateAPIKeyResponse struct {
	*apikey.Key
	Raw string `json:"key"`
}

// EndpointCreateAPIKey handles the 'POST /v1/api_keys' endpoint
func (service *Service) EndpointCreateAPIKey(writer http.ResponseWriter, request *http.Request) {
	payload, validationErrs, err := schema.UnmarshalBody[endpointCreateAPIKeyRequestPayload](request)
	if err != nil {
		service.writer.WriteInternalError(writer, err)
		return
	}
	if len(validationErrs) > 0 {
		service.writer.WriteErrors(writer, http.StatusBadRequest, validationErrs...)
		return
	}
	if *payload.Quota < 0 {
		*payload.Quota = -1
	}
	if *payload.RateLimit < 0 {
		*payload.RateLimit = -1
	}

	client := request.Context().Value(contextValueUser).(*user.User)
	if !client.Admin {
		var policyErrs []*schema.Error

		if !client.APIKeyPolicy.ValidateQuota(*payload.Quota) {
			policyErrs = append(policyErrs, errAPIKeyQuotaNotAllowed(*payload.Quota, client.APIKeyPolicy.MaxQuota))
		}
		if !client.APIKeyPolicy.ValidateRateLimit(*payload.RateLimit) {
			policyErrs = append(policyErrs, errAPIKeyRateLimitNotAllowed(*payload.RateLimit, client.APIKeyPolicy.MaxRateLimit))
		}
		if !client.APIKeyPolicy.ValidateCapabilities(*payload.Capabilities) {
			policyErrs = append(policyErrs, errAPIKeyCapabilitiesNotAllowed(*payload.Capabilities, client.APIKeyPolicy.AllowedCapabilities))
		}

		if len(policyErrs) > 0 {
			service.writer.WriteErrors(writer, http.StatusForbidden, policyErrs...)
			return
		}
	}

	create := &apikey.Create{
		UserID:       client.ID,
		Description:  "",
		Quota:        *payload.Quota,
		RateLimit:    *payload.RateLimit,
		Capabilities: *payload.Capabilities,
	}
	if payload.Description != nil {
		create.Description = apikey.SanitizeDescription(*payload.Description)
	}

	key, raw, err := service.Storage.APIKeys().Create(request.Context(), create)
	if err != nil {
		service.writer.WriteInternalError(writer, err)
		return
	}
	service.writer.WriteJSONWithCode(writer, http.StatusCreated, endpointCreateAPIKeyResponse{
		Key: key,
		Raw: raw,
	})
}

// EndpointGetAPIKeys handles the 'GET /v1/api_keys?offset={number?:0}&limit={number?:10}&user_id={string?}' endpoint
func (service *Service) EndpointGetAPIKeys(writer http.ResponseWriter, request *http.Request) {
	var validationErrs []*schema.Error

	offset, validationErr := schema.QueryNumber(request, "offset", false, 0, 0, math.MaxInt64)
	if validationErr != nil {
		validationErrs = append(validationErrs, validationErr)
	}

	limit, validationErr := schema.QueryNumber(request, "limit", false, 10, 1, 1000)
	if validationErr != nil {
		validationErrs = append(validationErrs, validationErr)
	}

	if len(validationErrs) > 0 {
		service.writer.WriteErrors(writer, http.StatusBadRequest, validationErrs...)
		return
	}

	userID := request.URL.Query().Get("user_id")

	client := request.Context().Value(contextValueUser).(*user.User)
	if !client.Admin && userID != client.ID {
		service.writer.WriteErrors(writer, http.StatusForbidden, schema.ErrForbidden)
		return
	}

	var keys []*apikey.Key
	var n uint64
	var err error
	if userID == "" {
		keys, n, err = service.Storage.APIKeys().Get(request.Context(), uint64(offset), uint64(limit))
	} else {
		keys, n, err = service.Storage.APIKeys().GetByUserID(request.Context(), userID, uint64(offset), uint64(limit))
	}
	if err != nil {
		service.writer.WriteInternalError(writer, err)
		return
	}

	service.writer.WriteJSON(writer, schema.BuildPaginatedResponse(uint64(offset), uint64(limit), n, keys))
}

// EndpointGetAPIKey handles the 'GET /v1/api_keys/{id}' endpoint
func (service *Service) EndpointGetAPIKey(writer http.ResponseWriter, request *http.Request) {
	client := request.Context().Value(contextValueUser).(*user.User)

	id := chi.URLParam(request, "id")
	uid, err := uuid.Parse(id)
	if err != nil {
		if client.Admin {
			service.writer.WriteErrors(writer, http.StatusNotFound, schema.ErrNotFound)
		} else {
			service.writer.WriteErrors(writer, http.StatusForbidden, schema.ErrForbidden)
		}
		return
	}

	obj, err := service.Storage.APIKeys().GetByID(request.Context(), uid)
	if err != nil {
		service.writer.WriteInternalError(writer, err)
		return
	}

	if !client.Admin && (obj == nil || obj.UserID != client.ID) {
		service.writer.WriteErrors(writer, http.StatusForbidden, schema.ErrForbidden)
		return
	}

	if obj == nil {
		service.writer.WriteErrors(writer, http.StatusNotFound, schema.ErrNotFound)
		return
	}

	service.writer.WriteJSON(writer, obj)
}

type endpointEditAPIKeyRequestPayload struct {
	Description  *string            `json:"description"`
	Quota        *int64             `json:"quota"`
	RateLimit    *int               `json:"rate_limit"`
	Capabilities *bitflag.Container `json:"capabilities"`
}

// EndpointEditAPIKey handles the 'PATCH /v1/api_keys/{id}' endpoint
func (service *Service) EndpointEditAPIKey(writer http.ResponseWriter, request *http.Request) {
	client := request.Context().Value(contextValueUser).(*user.User)

	id := chi.URLParam(request, "id")
	uid, err := uuid.Parse(id)
	if err != nil {
		if client.Admin {
			service.writer.WriteErrors(writer, http.StatusNotFound, schema.ErrNotFound)
		} else {
			service.writer.WriteErrors(writer, http.StatusForbidden, schema.ErrForbidden)
		}
		return
	}

	obj, err := service.Storage.APIKeys().GetByID(request.Context(), uid)
	if err != nil {
		service.writer.WriteInternalError(writer, err)
		return
	}

	if !client.Admin && (obj == nil || obj.UserID != client.ID) {
		service.writer.WriteErrors(writer, http.StatusForbidden, schema.ErrForbidden)
		return
	}

	if obj == nil {
		service.writer.WriteErrors(writer, http.StatusNotFound, schema.ErrNotFound)
		return
	}

	payload, validationErrs, err := schema.UnmarshalBody[endpointEditAPIKeyRequestPayload](request)
	if err != nil {
		service.writer.WriteInternalError(writer, err)
		return
	}
	if len(validationErrs) > 0 {
		service.writer.WriteErrors(writer, http.StatusBadRequest, validationErrs...)
		return
	}
	if payload.Quota != nil && *payload.Quota < 0 {
		*payload.Quota = -1
	}
	if payload.RateLimit != nil && *payload.RateLimit < 0 {
		*payload.RateLimit = -1
	}

	if !client.Admin {
		var policyErrs []*schema.Error

		if payload.Quota != nil && !client.APIKeyPolicy.ValidateQuota(*payload.Quota) {
			policyErrs = append(policyErrs, errAPIKeyQuotaNotAllowed(*payload.Quota, client.APIKeyPolicy.MaxQuota))
		}
		if payload.RateLimit != nil && !client.APIKeyPolicy.ValidateRateLimit(*payload.RateLimit) {
			policyErrs = append(policyErrs, errAPIKeyRateLimitNotAllowed(*payload.RateLimit, client.APIKeyPolicy.MaxRateLimit))
		}
		if payload.Capabilities != nil && !client.APIKeyPolicy.ValidateCapabilities(*payload.Capabilities) {
			policyErrs = append(policyErrs, errAPIKeyCapabilitiesNotAllowed(*payload.Capabilities, client.APIKeyPolicy.AllowedCapabilities))
		}

		if len(policyErrs) > 0 {
			service.writer.WriteErrors(writer, http.StatusForbidden, policyErrs...)
			return
		}
	}

	update := &apikey.Update{
		Quota:        payload.Quota,
		RateLimit:    payload.RateLimit,
		Capabilities: payload.Capabilities,
	}
	if payload.Description != nil {
		desc := apikey.SanitizeDescription(*payload.Description)
		update.Description = &desc
	}

	newObj, err := service.Storage.APIKeys().Update(request.Context(), obj.ID, update)
	if err != nil {
		service.writer.WriteInternalError(writer, err)
		return
	}
	service.writer.WriteJSONWithCode(writer, http.StatusOK, newObj)
}

// EndpointDeleteAPIKey handles the 'DELETE /v1/api_keys/{id}' endpoint
func (service *Service) EndpointDeleteAPIKey(writer http.ResponseWriter, request *http.Request) {
	client := request.Context().Value(contextValueUser).(*user.User)

	id := chi.URLParam(request, "id")
	uid, err := uuid.Parse(id)
	if err != nil {
		if client.Admin {
			service.writer.WriteErrors(writer, http.StatusNotFound, schema.ErrNotFound)
		} else {
			service.writer.WriteErrors(writer, http.StatusForbidden, schema.ErrForbidden)
		}
		return
	}

	obj, err := service.Storage.APIKeys().GetByID(request.Context(), uid)
	if err != nil {
		service.writer.WriteInternalError(writer, err)
		return
	}

	if !client.Admin && (obj == nil || obj.UserID != client.ID) {
		service.writer.WriteErrors(writer, http.StatusForbidden, schema.ErrForbidden)
		return
	}

	if obj == nil {
		service.writer.WriteErrors(writer, http.StatusNotFound, schema.ErrNotFound)
		return
	}

	if err := service.Storage.APIKeys().Delete(request.Context(), obj.ID); err != nil {
		service.writer.WriteInternalError(writer, err)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}

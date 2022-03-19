package portal

import (
	"github.com/go-chi/chi/v5"
	"github.com/skybi/data-server/internal/api/schema"
	"github.com/skybi/data-server/internal/api/validation"
	"github.com/skybi/data-server/internal/bitflag"
	"github.com/skybi/data-server/internal/user"
	"math"
	"net/http"
)

// EndpointGetUsers handles the 'GET /v1/users?offset={number?:0}&limit={number?:10}' endpoint
func (service *Service) EndpointGetUsers(writer http.ResponseWriter, request *http.Request) {
	var validationErrs []*schema.Error

	offset, validationErr := validation.QueryNumber(request, "offset", false, 0, 0, math.MaxInt64)
	if validationErr != nil {
		validationErrs = append(validationErrs, validationErr)
	}

	limit, validationErr := validation.QueryNumber(request, "limit", false, 10, 1, 1000)
	if validationErr != nil {
		validationErrs = append(validationErrs, validationErr)
	}

	if len(validationErrs) > 0 {
		service.writer.WriteErrors(writer, http.StatusBadRequest, validationErrs...)
		return
	}

	users, n, err := service.Storage.Users().Get(request.Context(), uint64(offset), uint64(limit))
	if err != nil {
		service.writer.WriteInternalError(writer, err)
		return
	}

	service.writer.WriteJSON(writer, schema.BuildPaginatedResponse(uint64(offset), uint64(limit), n, users))
}

// EndpointGetUser handles the 'GET /v1/users/{id}' endpoint
func (service *Service) EndpointGetUser(writer http.ResponseWriter, request *http.Request) {
	id := chi.URLParam(request, "id")

	obj, err := service.Storage.Users().GetByID(request.Context(), id)
	if err != nil {
		service.writer.WriteInternalError(writer, err)
		return
	}
	if obj == nil {
		service.writer.WriteErrors(writer, http.StatusNotFound, schema.ErrNotFound)
		return
	}

	service.writer.WriteJSON(writer, obj)
}

type endpointEditUserRequestPayload struct {
	Restricted   *bool `json:"restricted"`
	Admin        *bool `json:"admin"`
	APIKeyPolicy *struct {
		MaxQuota            *int64             `json:"max_quota"`
		MaxRateLimit        *int               `json:"max_rate_limit"`
		AllowedCapabilities *bitflag.Container `json:"allowed_capabilities"`
	} `json:"api_key_policy"`
}

// EndpointEditUser handles the 'PATCH /v1/users/{id}' endpoint
func (service *Service) EndpointEditUser(writer http.ResponseWriter, request *http.Request) {
	id := chi.URLParam(request, "id")

	// Retrieve the old user
	obj, err := service.Storage.Users().GetByID(request.Context(), id)
	if err != nil {
		service.writer.WriteInternalError(writer, err)
		return
	}
	if obj == nil {
		service.writer.WriteErrors(writer, http.StatusNotFound, schema.ErrNotFound)
		return
	}

	// Unmarshal and validate the request body
	payload, validationErrs, err := validation.UnmarshalBody[endpointEditUserRequestPayload](request)
	if err != nil {
		service.writer.WriteInternalError(writer, err)
		return
	}
	if len(validationErrs) > 0 {
		service.writer.WriteErrors(writer, http.StatusBadRequest, validationErrs...)
		return
	}

	// Construct the update action
	update := &user.Update{
		Restricted: payload.Restricted,
		Admin:      payload.Admin,
	}
	if payload.APIKeyPolicy != nil {
		update.APIKeyPolicy = &user.APIKeyPolicyUpdate{
			MaxQuota:            payload.APIKeyPolicy.MaxQuota,
			MaxRateLimit:        payload.APIKeyPolicy.MaxRateLimit,
			AllowedCapabilities: payload.APIKeyPolicy.AllowedCapabilities,
		}
	}

	// Update the user and return the new one
	newObj, err := service.Storage.Users().Update(request.Context(), obj.ID, update)
	if err != nil {
		service.writer.WriteInternalError(writer, err)
		return
	}
	service.writer.WriteJSON(writer, newObj)
}

// EndpointDeleteUserData handles the 'DELETE /v1/users/{id}' endpoint
func (service *Service) EndpointDeleteUserData(writer http.ResponseWriter, request *http.Request) {
	id := chi.URLParam(request, "id")

	obj, err := service.Storage.Users().GetByID(request.Context(), id)
	if err != nil {
		service.writer.WriteInternalError(writer, err)
		return
	}
	if obj == nil {
		service.writer.WriteErrors(writer, http.StatusNotFound, schema.ErrNotFound)
		return
	}

	if err := service.Storage.Users().Delete(request.Context(), obj.ID); err != nil {
		service.writer.WriteInternalError(writer, err)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}

// EndpointGetSelfUser handles the 'GET /v1/me' endpoint
func (service *Service) EndpointGetSelfUser(writer http.ResponseWriter, request *http.Request) {
	obj := request.Context().Value(contextValueUser).(*user.User)
	service.writer.WriteJSON(writer, obj)
}

// EndpointDeleteSelfUserData handles the 'DELETE /v1/me' endpoint
func (service *Service) EndpointDeleteSelfUserData(writer http.ResponseWriter, request *http.Request) {
	obj := request.Context().Value(contextValueUser).(*user.User)
	if err := service.Storage.Users().Delete(request.Context(), obj.ID); err != nil {
		service.writer.WriteInternalError(writer, err)
		return
	}
	unsetCookie(writer, sessionTokenCookieName)
	if err := service.sessionStorage.TerminateByUserID(request.Context(), obj.ID); err != nil {
		service.writer.WriteInternalError(writer, err)
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

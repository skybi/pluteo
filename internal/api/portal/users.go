package portal

import (
	"github.com/go-chi/chi/v5"
	"github.com/skybi/data-server/internal/api/schema"
	"github.com/skybi/data-server/internal/api/validation"
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

// EndpointGetSelfUser handles the 'GET /v1/me' endpoint
func (service *Service) EndpointGetSelfUser(writer http.ResponseWriter, request *http.Request) {
	obj := request.Context().Value(contextValueUser).(*user.User)
	service.writer.WriteJSON(writer, obj)
}

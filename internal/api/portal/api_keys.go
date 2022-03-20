package portal

import (
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/skybi/data-server/internal/api/schema"
	"github.com/skybi/data-server/internal/api/validation"
	"github.com/skybi/data-server/internal/apikey"
	"github.com/skybi/data-server/internal/user"
	"math"
	"net/http"
)

// EndpointGetAPIKeys handles the 'GET /v1/api_keys?offset={number?:0}&limit={number?:10}&user_id={string?}' endpoint
func (service *Service) EndpointGetAPIKeys(writer http.ResponseWriter, request *http.Request) {
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

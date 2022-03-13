package portal

import (
	"github.com/skybi/data-server/internal/api/schema"
	"github.com/skybi/data-server/internal/api/validation"
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

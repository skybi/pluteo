package data

import (
	"github.com/skybi/data-server/internal/api/schema"
	"github.com/skybi/data-server/internal/apikey"
	"github.com/skybi/data-server/internal/metar"
	"math"
	"net/http"
	"strings"
)

// EndpointGetMETARs handles the 'GET /v1/metars?station_id={string?}&before={timestamp?}&after={timestamp?}&limit={number?:10}' endpoint
func (service *Service) EndpointGetMETARs(writer http.ResponseWriter, request *http.Request) {
	var validationErrs []*schema.Error

	stationID := strings.ToUpper(strings.TrimSpace(request.URL.Query().Get("station_id")))

	before, validationErr := schema.QueryNumber(request, "before", false, -1, 0, math.MaxInt64)
	if validationErr != nil {
		validationErrs = append(validationErrs, validationErr)
	}

	after, validationErr := schema.QueryNumber(request, "after", false, -1, 0, math.MaxInt64)
	if validationErr != nil {
		validationErrs = append(validationErrs, validationErr)
	}

	limit, validationErr := schema.QueryNumber(request, "limit", false, 10, 1, 100)
	if validationErr != nil {
		validationErrs = append(validationErrs, validationErr)
	}

	if len(validationErrs) > 0 {
		service.writer.WriteErrors(writer, http.StatusBadRequest, validationErrs...)
		return
	}

	filter := &metar.Filter{}
	if stationID != "" {
		filter.StationID = &stationID
	}
	if before > 0 {
		filter.IssuedBefore = &before
	}
	if after > 0 {
		filter.IssuedAfter = &after
	}

	metars, n, err := service.Storage.METARs().GetByFilter(request.Context(), filter, uint64(limit))
	if err != nil {
		service.writer.WriteInternalError(writer, err)
		return
	}

	service.writer.WriteJSON(writer, schema.BuildPaginatedResponse(0, uint64(limit), n, metars))

	service.QuotaTracker.Accumulate(request.Context().Value(contextValueKey).(*apikey.Key))
}

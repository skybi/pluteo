package data

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/skybi/pluteo/internal/api/schema"
	"github.com/skybi/pluteo/internal/apikey"
	"github.com/skybi/pluteo/internal/metar"
	"math"
	"net/http"
	"strings"
)

var metarFeedBatchMaxSize = 500

var (
	errMETARTooLargeBatch = func(given, max int) *schema.Error {
		return &schema.Error{
			Type:    "schema.metars.tooLargeBatch",
			Message: fmt.Sprintf("A single METAR request may only feed %d METARs (%d were given).", max, given),
			Details: map[string]any{
				"given": given,
				"max":   max,
			},
		}
	}
	errMETARInvalidFormat = func(raw string, i int) *schema.Error {
		return &schema.Error{
			Type:    "data.metars.invalidFormat",
			Message: raw,
			Details: map[string]any{
				"index": i,
			},
		}
	}
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

// EndpointGetMETAR handles the 'GET /v1/metars/{id}' endpoint
func (service *Service) EndpointGetMETAR(writer http.ResponseWriter, request *http.Request) {
	id := chi.URLParam(request, "id")
	uid, err := uuid.Parse(id)
	if err != nil {
		service.writer.WriteErrors(writer, http.StatusNotFound, schema.ErrNotFound)
		return
	}

	obj, err := service.Storage.METARs().GetByID(request.Context(), uid)
	if err != nil {
		service.writer.WriteInternalError(writer, err)
		return
	}

	service.writer.WriteJSON(writer, obj)

	service.QuotaTracker.Accumulate(request.Context().Value(contextValueKey).(*apikey.Key))
}

type endpointFeedMETARsRequestPayload struct {
	Data []string `json:"data" required:"true"`
}

type endpointFeedMETARsResponseBody struct {
	METARs     []*metar.METAR `json:"metars"`
	Duplicates []uint         `json:"duplicates"`
}

// EndpointFeedMETARs handles the 'POST /v1/metars' endpoint
func (service *Service) EndpointFeedMETARs(writer http.ResponseWriter, request *http.Request) {
	body, validationErrs, err := schema.UnmarshalBody[endpointFeedMETARsRequestPayload](request)
	if len(validationErrs) > 0 {
		service.writer.WriteErrors(writer, http.StatusBadRequest, validationErrs...)
		return
	}
	if err != nil {
		service.writer.WriteInternalError(writer, err)
		return
	}

	if len(body.Data) > metarFeedBatchMaxSize {
		service.writer.WriteErrors(writer, http.StatusRequestEntityTooLarge, errMETARTooLargeBatch(len(body.Data), metarFeedBatchMaxSize))
		return
	}

	metars, duplicates, err := service.Storage.METARs().Create(request.Context(), body.Data)
	if err != nil {
		var formatErr *metar.FormatError
		if errors.As(err, &formatErr) {
			service.writer.WriteErrors(writer, http.StatusBadRequest, errMETARInvalidFormat(err.Error(), formatErr.Index))
		} else {
			service.writer.WriteInternalError(writer, err)
		}
		return
	}

	service.writer.WriteJSON(writer, endpointFeedMETARsResponseBody{
		METARs:     metars,
		Duplicates: duplicates,
	})
}

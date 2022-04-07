package data

import (
	"github.com/skybi/data-server/internal/apikey"
	"net/http"
)

// EndpointGetKeyInfo handles the 'GET /v1/key_info' endpoint
func (service *Service) EndpointGetKeyInfo(writer http.ResponseWriter, request *http.Request) {
	key := request.Context().Value(contextValueKey).(*apikey.Key)
	cpy := *key
	cpy.UsedQuota = service.QuotaTracker.Get(key)
	service.writer.WriteJSON(writer, cpy)
}

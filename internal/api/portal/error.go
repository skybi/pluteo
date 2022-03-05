package portal

import (
	"encoding/json"
	"github.com/rs/zerolog/log"
	"net/http"
)

func (service *Service) error(writer http.ResponseWriter, status int, message string) {
	response, _ := json.Marshal(map[string]interface{}{
		"status":  status,
		"message": message,
	})
	writer.WriteHeader(status)
	writer.Write(response)
}

func (service *Service) internalError(writer http.ResponseWriter, err error) {
	service.error(writer, http.StatusInternalServerError, "internal error")
	log.Error().Err(err).Msg("the portal API experienced an unexpected error")
}

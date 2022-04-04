package schema

import (
	"encoding/json"
	"net/http"
)

// Writer helps writing unified API responses
type Writer struct {
	InternalErrorHook func(err error)
}

// WriteJSONWithCode writes the JSON representation of value to the given response writer using the given HTTP status code
func (writer *Writer) WriteJSONWithCode(rw http.ResponseWriter, code int, value any) {
	val, _ := json.Marshal(value)
	rw.WriteHeader(code)
	rw.Write(val)
}

// WriteJSON writes the JSON representation of value to the given response writer.
// This method sends 200 OK as the HTTP status code; use WriteJSONWithCode to use a different one.
func (writer *Writer) WriteJSON(rw http.ResponseWriter, value any) {
	writer.WriteJSONWithCode(rw, http.StatusOK, value)
}

// WriteErrors sends an error response
func (writer *Writer) WriteErrors(rw http.ResponseWriter, code int, errors ...*Error) {
	if errors == nil {
		errors = []*Error{}
	}
	response := &ErrorResponse{
		Status: code,
		Errors: errors,
	}
	for _, err := range response.Errors {
		if err.Details == nil {
			err.Details = map[string]any{}
		}
	}
	writer.WriteJSONWithCode(rw, code, response)
}

// WriteInternalError processes an internal server error and writes it to the response
func (writer *Writer) WriteInternalError(rw http.ResponseWriter, err error) {
	writer.InternalErrorHook(err)
	writer.WriteErrors(rw, http.StatusInternalServerError, ErrInternal)
}

package schema

import (
	"fmt"
	"net/http"
	"strconv"
)

var (
	errQueryParameterMissing = func(name string) *Error {
		return &Error{
			Type:    "validation.query.parameter.missing",
			Message: fmt.Sprintf("The query parameter '%s' is required but was not present in the request.", name),
			Details: map[string]any{
				"parameter": name,
			},
		}
	}
	errQueryParameterInvalidType = func(name, value, expectedType string) *Error {
		return &Error{
			Type:    "validation.query.parameter.invalidType",
			Message: fmt.Sprintf("The query parameter '%s' ('%s') could not be assigned to the required type (%s).", name, value, expectedType),
			Details: map[string]any{
				"parameter":     name,
				"value":         value,
				"expected_type": expectedType,
			},
		}
	}
	errQueryParameterNumberOutOfRange = func(name string, value, min, max int64) *Error {
		comparison := ""
		if value < min {
			comparison = fmt.Sprintf("%d [given] < %d [min]", value, min)
		} else if value > max {
			comparison = fmt.Sprintf("%d [given] > %d [max]", value, max)
		}

		return &Error{
			Type:    "validation.query.parameter.number.outOfRange",
			Message: fmt.Sprintf("The query parameter '%s' is out of the required range (%s).", name, comparison),
			Details: map[string]any{
				"parameter": name,
				"value":     value,
				"min":       min,
				"max":       max,
			},
		}
	}
)

// QueryNumber extracts and validates an integer value out of the query parameters of the given request
func QueryNumber(request *http.Request, key string, required bool, def, min, max int64) (int64, *Error) {
	// Extract the raw string value
	value := request.URL.Query().Get(key)
	if value == "" {
		if required {
			return 0, errQueryParameterMissing(key)
		}
		return def, nil
	}

	// Try to parse the value
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, errQueryParameterInvalidType(key, value, "number")
	}

	// Check if the parsed value is in the required range
	if parsed < min || parsed > max {
		return 0, errQueryParameterNumberOutOfRange(key, parsed, min, max)
	}

	return parsed, nil
}

package schema

var emptyMap = map[string]interface{}{}

var (
	ErrInternal = &Error{
		Type:    "generic.internal",
		Message: "An internal error occurred.",
		Details: emptyMap,
	}
	ErrNotFound = &Error{
		Type:    "generic.notFound",
		Message: "Resource not found.",
		Details: emptyMap,
	}
	ErrMethodNotAllowed = &Error{
		Type:    "generic.methodNotAllowed",
		Message: "Method not allowed.",
		Details: emptyMap,
	}
	ErrUnauthorized = &Error{
		Type:    "generic.unauthorized",
		Message: "Unauthorized",
		Details: emptyMap,
	}
)

// ErrorResponse represents the response structure sent by the portal or data API whenever errors occurred
type ErrorResponse struct {
	Status int      `json:"status"`
	Errors []*Error `json:"errors"`
}

// Error represents a single error present in the ErrorResponse
type Error struct {
	Type    string                 `json:"type"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details"`
}

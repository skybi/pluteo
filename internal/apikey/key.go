package apikey

import (
	"github.com/google/uuid"
)

// Key represents an API key used to access the data API
type Key struct {
	Key          uuid.UUID    `json:"key,omitempty"`
	UserID       string       `json:"user_id"`
	Quota        int64        `json:"quota"`
	RateLimit    int          `json:"rate_limit"`
	Capabilities Capabilities `json:"capabilities"`
}

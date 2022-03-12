package apikey

import (
	"github.com/google/uuid"
)

// Key represents an API key used to access the data API
type Key struct {
	ID           uuid.UUID    `json:"id"`
	Key          string       `json:"key,omitempty"`
	UserID       string       `json:"user_id"`
	Description  string       `json:"description"`
	Quota        int64        `json:"quota"`
	UsedQuota    int64        `json:"used_quota"`
	RateLimit    int          `json:"rate_limit"`
	Capabilities Capabilities `json:"capabilities"`
}

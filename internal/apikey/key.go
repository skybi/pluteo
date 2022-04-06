package apikey

import (
	"github.com/google/uuid"
	"github.com/skybi/data-server/internal/bitflag"
	"strings"
	"unicode/utf8"
)

// MaxDescriptionLength defines the maximum length an API key description may have
var MaxDescriptionLength = 100

// SanitizeDescription sanitizes a description by turning it into a valid UTF8 string, trimming leading and trailing
// spaces, replacing newlines with spaces and stripping it to the maximum length a description may have
func SanitizeDescription(raw string) string {
	raw = strings.ToValidUTF8(raw, "?")
	raw = strings.TrimSpace(raw)
	raw = strings.ReplaceAll(raw, "\n", " ")

	if utf8.RuneCountInString(raw) > MaxDescriptionLength {
		return string([]rune(raw)[:MaxDescriptionLength])
	}
	return raw
}

// Key represents an API key used to access the data API
type Key struct {
	ID           uuid.UUID         `json:"id"`
	Key          []byte            `json:"-"`
	UserID       string            `json:"user_id"`
	Description  string            `json:"description"`
	Quota        int64             `json:"quota"`
	UsedQuota    int64             `json:"used_quota"`
	RateLimit    int               `json:"rate_limit"`
	Capabilities bitflag.Container `json:"capabilities"`
}

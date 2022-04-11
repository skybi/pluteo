package user

import (
	"github.com/skybi/pluteo/internal/apikey"
	"github.com/skybi/pluteo/internal/bitflag"
)

// DefaultAPIKeyPolicy returns the default API key policy to use for new users
func DefaultAPIKeyPolicy() *APIKeyPolicy {
	return &APIKeyPolicy{
		MaxQuota:     -1, // Unlimited requests
		MaxRateLimit: 60, // Max. 60 requests per minute
		AllowedCapabilities: bitflag.EmptyContainer.With(
			apikey.CapabilityReadMETARs,
		),
	}
}

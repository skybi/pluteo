package user

import "github.com/skybi/data-server/internal/bitflag"

// DefaultAPIKeyPolicy returns the default API key policy to use for new users
func DefaultAPIKeyPolicy() *APIKeyPolicy {
	return &APIKeyPolicy{
		MaxQuota:            0,                      // Unlimited requests
		MaxRateLimit:        60,                     // Max. 60 requests per minute
		AllowedCapabilities: bitflag.EmptyContainer, // TODO: Update
	}
}

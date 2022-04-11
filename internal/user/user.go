package user

import (
	"github.com/skybi/pluteo/internal/bitflag"
)

// User represents a user registered to the service
type User struct {
	ID           string        `json:"id"`
	DisplayName  string        `json:"display_name"`
	APIKeyPolicy *APIKeyPolicy `json:"api_key_policy,omitempty"`
	Restricted   bool          `json:"restricted"`
	Admin        bool          `json:"admin"`
}

// APIKeyPolicy represents the user-specific policy to create API keys
type APIKeyPolicy struct {
	MaxQuota            int64             `json:"max_quota"`
	MaxRateLimit        int               `json:"max_rate_limit"`
	AllowedCapabilities bitflag.Container `json:"allowed_capabilities"`
}

// ValidateQuota checks if the given quota is allowed as defined by the API key policy
func (policy *APIKeyPolicy) ValidateQuota(quota int64) bool {
	return policy.MaxQuota < 0 || policy.MaxQuota >= quota
}

// ValidateRateLimit checks if the given rate limit is allowed as defined by the API key policy
func (policy *APIKeyPolicy) ValidateRateLimit(rateLimit int) bool {
	return policy.MaxRateLimit < 0 || policy.MaxRateLimit >= rateLimit
}

// ValidateCapabilities checks if the given capabilities are allowed as defined by the API key policy
func (policy *APIKeyPolicy) ValidateCapabilities(capabilities bitflag.Container) bool {
	allowed := uint(policy.AllowedCapabilities)
	compare := uint(capabilities)
	return allowed&compare == compare
}

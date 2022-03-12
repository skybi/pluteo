package user

import (
	"context"
	"github.com/skybi/data-server/internal/apikey"
)

// Repository defines the user repository API
type Repository interface {
	// GetByID retrieves a user by their ID
	GetByID(ctx context.Context, id string) (*User, error)

	// Create creates a new user
	Create(ctx context.Context, create *Create) (*User, error)

	// Update updates an existing user
	Update(ctx context.Context, id string, update *Update) (*User, error)

	// Delete deletes a user by their ID
	Delete(ctx context.Context, id string) error
}

// Create is used to create a new user
type Create struct {
	ID           string
	DisplayName  string
	APIKeyPolicy *APIKeyPolicy
}

// Update is used to update an existing user
type Update struct {
	DisplayName  *string
	APIKeyPolicy *APIKeyPolicyUpdate
	Restricted   *bool
}

// APIKeyPolicyUpdate is used to update the API key policy of an existing user
type APIKeyPolicyUpdate struct {
	MaxQuota            *int64
	MaxRateLimit        *int
	AllowedCapabilities *apikey.Capabilities
}

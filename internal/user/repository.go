package user

import (
	"context"
	"github.com/skybi/data-server/internal/bitflag"
)

// Repository defines the user repository API
type Repository interface {
	// Get retrieves multiple users
	Get(ctx context.Context, offset, limit uint64) ([]*User, uint64, error)

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
	Admin        bool
}

// Update is used to update an existing user
type Update struct {
	DisplayName  *string
	APIKeyPolicy *APIKeyPolicyUpdate
	Restricted   *bool
	Admin        *bool
}

// APIKeyPolicyUpdate is used to update the API key policy of an existing user
type APIKeyPolicyUpdate struct {
	MaxQuota            *int64
	MaxRateLimit        *int
	AllowedCapabilities *bitflag.Container
}

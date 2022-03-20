package apikey

import (
	"context"
	"github.com/google/uuid"
	"github.com/skybi/data-server/internal/bitflag"
)

// Repository defines the API key repository API
type Repository interface {
	// Get retrieves multiple API keys
	Get(ctx context.Context, offset, limit uint64) ([]*Key, uint64, error)

	// GetByUserID retrieves multiple API keys of a specific user
	GetByUserID(ctx context.Context, userID string, offset, limit uint64) ([]*Key, uint64, error)

	// GetByID retrieves an API key by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*Key, error)

	// GetByRawKey retrieves an API key by the raw bearer token
	GetByRawKey(ctx context.Context, key string) (*Key, error)

	// Create creates a new API key
	Create(ctx context.Context, create *Create) (*Key, error)

	// Update updates an API key
	Update(ctx context.Context, id uuid.UUID, update *Update) (*Key, error)

	// Delete deletes an API key by its ID
	Delete(ctx context.Context, id uuid.UUID) error
}

// Create is used to create a new API key
type Create struct {
	UserID       string
	Description  string
	Quota        int64
	RateLimit    int
	Capabilities bitflag.Container
}

// Update is used to update an existing API key
type Update struct {
	Description  *string
	Quota        *int64
	UsedQuota    *int64
	RateLimit    *int
	Capabilities *bitflag.Container
}

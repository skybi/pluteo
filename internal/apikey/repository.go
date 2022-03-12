package apikey

import (
	"context"
	"github.com/google/uuid"
)

// Repository defines the API key repository API
type Repository interface {
	// GetByID retrieves an API key by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*Key, error)

	// GetByRawKey retrieves an API key by the raw bearer token
	GetByRawKey(ctx context.Context, key string) (*Key, error)

	// GetByUserID retrieves all API keys of a specific user
	GetByUserID(ctx context.Context, userID string) ([]*Key, error)

	// Create creates a new API key.
	// This method may hash the key field of the given key.
	Create(ctx context.Context, key *Key) error

	// Delete deletes an API key by its ID
	Delete(ctx context.Context, id uuid.UUID) error
}

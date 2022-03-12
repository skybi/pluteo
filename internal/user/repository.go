package user

import "context"

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
	APIKeyPolicy *APIKeyPolicy
	Restricted   *bool
}

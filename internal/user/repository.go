package user

import "context"

// Repository defines the user repository API
type Repository interface {
	// GetByID retrieves a user by their ID
	GetByID(ctx context.Context, id string) (*User, error)

	// Create creates a new user
	Create(ctx context.Context, user *User) error

	// Delete deletes a user by their ID
	Delete(ctx context.Context, id string) error
}

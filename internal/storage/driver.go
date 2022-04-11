package storage

import (
	"context"
	"github.com/skybi/pluteo/internal/apikey"
	"github.com/skybi/pluteo/internal/metar"
	"github.com/skybi/pluteo/internal/user"
)

// Driver represents a storage driver
type Driver interface {
	// Initialize initializes the storage driver (i.e. opens a database connection)
	Initialize(ctx context.Context) error

	// Users provides a user repository implementation
	Users() user.Repository

	// APIKeys provides an API key repository implementation
	APIKeys() apikey.Repository

	// METARs provides an API key repository implementation
	METARs() metar.Repository

	// Close closes the storage driver (i.e. closes a database connection)
	Close()
}

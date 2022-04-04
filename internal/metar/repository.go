package metar

import (
	"context"
	"github.com/google/uuid"
)

// Repository defines the METAR repository API
type Repository interface {
	// GetByFilter retrieves multiple METARs following a filter, ordered by their issuing date (descending).
	// If limit <= 0, a default limit value of 10 is used.
	GetByFilter(ctx context.Context, filter *Filter, limit uint64) ([]*METAR, uint64, error)

	// GetByID retrieves a METAR by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*METAR, error)

	// Create creates new METARs based on their raw text representation.
	// All raw strings are sanitized (leading and trailing spaces are trimmed).
	// This method also returns the indexes of the METARs that already exist in the database and thus were not inserted.
	Create(ctx context.Context, raw []string) ([]*METAR, []uint, error)

	// Delete deletes a METAR by its ID
	Delete(ctx context.Context, id uuid.UUID) error
}

// Filter is used to query METARs based on a filter
type Filter struct {
	StationID    *string
	IssuedBefore *int64
	IssuedAfter  *int64
}

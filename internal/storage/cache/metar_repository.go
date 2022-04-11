package cache

import (
	"context"
	"github.com/google/uuid"
	"github.com/skybi/pluteo/internal/hashmap"
	"github.com/skybi/pluteo/internal/metar"
)

// METARRepository implements the metar.Repository interface in order to implement caching
type METARRepository struct {
	repo  metar.Repository
	cache *hashmap.ExpiringMap[uuid.UUID, *metar.METAR]
}

var _ metar.Repository = (*METARRepository)(nil)

// GetByFilter retrieves multiple METARs following a filter, ordered by their issuing date (descending).
// If limit <= 0, a default limit value of 10 is used.
func (repo *METARRepository) GetByFilter(ctx context.Context, filter *metar.Filter, limit uint64) ([]*metar.METAR, uint64, error) {
	metars, n, err := repo.repo.GetByFilter(ctx, filter, limit)
	if err != nil {
		return nil, 0, err
	}
	for _, obj := range metars {
		repo.cache.Set(obj.ID, obj)
	}
	return metars, n, nil
}

// GetByID retrieves a METAR by its ID
func (repo *METARRepository) GetByID(ctx context.Context, id uuid.UUID) (*metar.METAR, error) {
	cached, ok := repo.cache.Lookup(id)
	if ok {
		return cached, nil
	}
	obj, err := repo.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if obj != nil {
		repo.cache.Set(obj.ID, obj)
	}
	return obj, nil
}

// Create creates new METARs based on their raw text representation.
// All raw strings are sanitized (leading and trailing spaces are trimmed).
// This method also returns the indexes of the METARs that already exist in the database and thus were not inserted.
func (repo *METARRepository) Create(ctx context.Context, raw []string) ([]*metar.METAR, []uint, error) {
	metars, duplicates, err := repo.repo.Create(ctx, raw)
	if err != nil {
		return nil, nil, err
	}
	for _, obj := range metars {
		repo.cache.Set(obj.ID, obj)
	}
	return metars, duplicates, nil
}

// Delete deletes a METAR by its ID
func (repo *METARRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := repo.repo.Delete(ctx, id)
	if err != nil {
		return err
	}
	repo.cache.Unset(id)
	return nil
}

package cache

import (
	"context"
	"github.com/google/uuid"
	"github.com/skybi/data-server/internal/apikey"
	"github.com/skybi/data-server/internal/hashmap"
	"github.com/skybi/data-server/internal/secret"
)

// APIKeyRepository implements the apikey.Repository interface in order to implement caching
type APIKeyRepository struct {
	repo      apikey.Repository
	cache     *hashmap.ExpiringMap[uuid.UUID, *apikey.Key]
	hashCache *hashmap.ExpiringMap[[64]byte, uuid.UUID]
}

var _ apikey.Repository = (*APIKeyRepository)(nil)

// Get retrieves multiple API keys
func (repo *APIKeyRepository) Get(ctx context.Context, offset, limit uint64) ([]*apikey.Key, uint64, error) {
	keys, n, err := repo.repo.Get(ctx, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	for _, key := range keys {
		repo.cache.Set(key.ID, key)
	}
	return keys, n, nil
}

// GetByUserID retrieves multiple API keys of a specific user
func (repo *APIKeyRepository) GetByUserID(ctx context.Context, userID string, offset, limit uint64) ([]*apikey.Key, uint64, error) {
	keys, n, err := repo.repo.GetByUserID(ctx, userID, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	for _, key := range keys {
		repo.cache.Set(key.ID, key)
	}
	return keys, n, nil
}

// GetByID retrieves an API key by its ID
func (repo *APIKeyRepository) GetByID(ctx context.Context, id uuid.UUID) (*apikey.Key, error) {
	cached, ok := repo.cache.Lookup(id)
	if ok {
		return cached, nil
	}
	key, err := repo.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if key != nil {
		repo.cache.Set(key.ID, key)
	}
	return key, nil
}

// GetByRawKey retrieves an API key by the raw bearer token
func (repo *APIKeyRepository) GetByRawKey(ctx context.Context, key string) (*apikey.Key, error) {
	hash, err := secret.Hash(key)
	if err != nil {
		return nil, err
	}
	id, ok := repo.hashCache.Lookup(hash)
	if ok {
		return repo.GetByID(ctx, id)
	}

	obj, err := repo.repo.GetByRawKey(ctx, key)
	if err != nil {
		return nil, err
	}
	if obj != nil {
		repo.hashCache.Set(hash, obj.ID)
		repo.cache.Set(obj.ID, obj)
	}
	return obj, nil
}

// Create creates a new API key
func (repo *APIKeyRepository) Create(ctx context.Context, create *apikey.Create) (*apikey.Key, string, error) {
	key, raw, err := repo.repo.Create(ctx, create)
	if err != nil {
		return nil, "", err
	}
	repo.cache.Set(key.ID, key)
	return key, raw, nil
}

// Update updates an API key
func (repo *APIKeyRepository) Update(ctx context.Context, id uuid.UUID, update *apikey.Update) (*apikey.Key, error) {
	key, err := repo.repo.Update(ctx, id, update)
	if err != nil {
		return nil, err
	}
	repo.cache.Set(key.ID, key)
	return key, nil
}

// UpdateManyQuotas updates many used API quotas at once
func (repo *APIKeyRepository) UpdateManyQuotas(ctx context.Context, updates map[uuid.UUID]int64) error {
	return repo.repo.UpdateManyQuotas(ctx, updates)
}

// Delete deletes an API key by its ID
func (repo *APIKeyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := repo.repo.Delete(ctx, id)
	if err != nil {
		return err
	}
	repo.cache.Unset(id)
	return nil
}

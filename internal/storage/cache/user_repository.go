package cache

import (
	"context"
	"github.com/skybi/data-server/internal/hashmap"
	"github.com/skybi/data-server/internal/user"
)

// UserRepository implements the user.Repository interface in order to implement caching
type UserRepository struct {
	repo  user.Repository
	cache *hashmap.ExpiringMap[string, *user.User]
}

var _ user.Repository = (*UserRepository)(nil)

// Get retrieves multiple users
func (repo *UserRepository) Get(ctx context.Context, offset, limit uint64) ([]*user.User, uint64, error) {
	users, n, err := repo.repo.Get(ctx, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	for _, obj := range users {
		repo.cache.Set(obj.ID, obj)
	}
	return users, n, nil
}

// GetByID retrieves a user by their ID
func (repo *UserRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
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

// Create creates a new user
func (repo *UserRepository) Create(ctx context.Context, create *user.Create) (*user.User, error) {
	obj, err := repo.repo.Create(ctx, create)
	if err != nil {
		return nil, err
	}
	repo.cache.Set(obj.ID, obj)
	return obj, nil
}

// Update updates an existing user
func (repo *UserRepository) Update(ctx context.Context, id string, update *user.Update) (*user.User, error) {
	obj, err := repo.repo.Update(ctx, id, update)
	if err != nil {
		return nil, err
	}
	repo.cache.Set(obj.ID, obj)
	return obj, nil
}

// Delete deletes a user by their ID
func (repo *UserRepository) Delete(ctx context.Context, id string) error {
	err := repo.repo.Delete(ctx, id)
	if err != nil {
		return err
	}
	repo.cache.Unset(id)
	return nil
}

package cache

import (
	"context"
	"github.com/skybi/data-server/internal/apikey"
	"github.com/skybi/data-server/internal/hashmap"
	"github.com/skybi/data-server/internal/metar"
	"github.com/skybi/data-server/internal/storage"
	"github.com/skybi/data-server/internal/user"
	"time"
)

// Driver represents a storage driver implementation that wraps another one in order to implement in-memory caching
type Driver struct {
	underlying storage.Driver
	users      *UserRepository
}

var _ storage.Driver = (*Driver)(nil)

// New returns a new caching storage driver
func New(underlying storage.Driver) *Driver {
	return &Driver{
		underlying: underlying,
	}
}

// Initialize initializes the caching repositories
func (driver *Driver) Initialize(_ context.Context) error {
	userCache := hashmap.NewExpiring[string, *user.User](5 * time.Minute)
	userCache.ScheduleCleanupTask(10 * time.Second)
	driver.users = &UserRepository{
		repo:  driver.underlying.Users(),
		cache: userCache,
	}

	return nil
}

// Users provides the caching user repository implementation
func (driver *Driver) Users() user.Repository {
	return driver.users
}

// APIKeys provides the caching API key repository implementation
func (driver *Driver) APIKeys() apikey.Repository {
	// TODO implement me
	return driver.underlying.APIKeys()
}

// METARs provides the caching METAR repository implementation
func (driver *Driver) METARs() metar.Repository {
	//TODO implement me
	return driver.underlying.METARs()
}

// Close closes the caching repositories and disposes their instances
func (driver *Driver) Close() {
	driver.users.cache.StopCleanupTask()
	driver.users = nil
}

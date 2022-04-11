package cache

import (
	"context"
	"github.com/google/uuid"
	"github.com/skybi/pluteo/internal/apikey"
	"github.com/skybi/pluteo/internal/hashmap"
	"github.com/skybi/pluteo/internal/metar"
	"github.com/skybi/pluteo/internal/storage"
	"github.com/skybi/pluteo/internal/user"
	"time"
)

// Driver represents a storage driver implementation that wraps another one in order to implement in-memory caching
type Driver struct {
	underlying storage.Driver
	users      *UserRepository
	apiKeys    *APIKeyRepository
	metars     *METARRepository
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
	userCache.ScheduleCleanupTask(time.Minute)
	driver.users = &UserRepository{
		repo:  driver.underlying.Users(),
		cache: userCache,
	}

	apiKeyCache := hashmap.NewExpiring[uuid.UUID, *apikey.Key](5 * time.Minute)
	apiKeyCache.ScheduleCleanupTask(time.Minute)
	apiKeyHashCache := hashmap.NewExpiring[[64]byte, uuid.UUID](5 * time.Minute)
	apiKeyHashCache.ScheduleCleanupTask(time.Minute)
	driver.apiKeys = &APIKeyRepository{
		repo:      driver.underlying.APIKeys(),
		cache:     apiKeyCache,
		hashCache: apiKeyHashCache,
	}

	metarCache := hashmap.NewExpiring[uuid.UUID, *metar.METAR](5 * time.Minute)
	metarCache.ScheduleCleanupTask(time.Minute)
	driver.metars = &METARRepository{
		repo:  driver.underlying.METARs(),
		cache: metarCache,
	}

	return nil
}

// Users provides the caching user repository implementation
func (driver *Driver) Users() user.Repository {
	return driver.users
}

// APIKeys provides the caching API key repository implementation
func (driver *Driver) APIKeys() apikey.Repository {
	return driver.apiKeys
}

// METARs provides the caching METAR repository implementation
func (driver *Driver) METARs() metar.Repository {
	return driver.metars
}

// Close closes the caching repositories and disposes their instances
func (driver *Driver) Close() {
	driver.users.cache.StopCleanupTask()
	driver.users = nil
	driver.apiKeys.cache.StopCleanupTask()
	driver.apiKeys.hashCache.StopCleanupTask()
	driver.apiKeys = nil
	driver.metars.cache.StopCleanupTask()
	driver.metars = nil
}

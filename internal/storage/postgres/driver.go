package postgres

import (
	"context"
	"embed"
	"errors"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/skybi/data-server/internal/apikey"
	"github.com/skybi/data-server/internal/metar"
	"github.com/skybi/data-server/internal/storage"
	"github.com/skybi/data-server/internal/user"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Driver represents the PostgreSQL storage driver implementation
type Driver struct {
	dsn     string
	db      *pgxpool.Pool
	users   *UserRepository
	apiKeys *APIKeyRepository
	metars  *METARRepository
}

var _ storage.Driver = (*Driver)(nil)

// New creates a new empty PostgreSQL storage driver.
// User Initialize to open the database connection and initialize the repository implementations.
func New(dsn string) *Driver {
	return &Driver{
		dsn: dsn,
	}
}

// Initialize opens the database connection, migrates the database and initializes the repository implementations
func (driver *Driver) Initialize(ctx context.Context) error {
	// Perform SQL migrations
	source, err := iofs.New(migrations, "migrations")
	if err != nil {
		return err
	}
	migrator, err := migrate.NewWithSourceInstance("iofs", source, driver.dsn)
	if err != nil {
		return err
	}
	defer migrator.Close()
	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	// Initialize the database connection pool
	pool, err := pgxpool.Connect(ctx, driver.dsn)
	if err != nil {
		return err
	}
	driver.db = pool

	// Initialize the repository implementations
	driver.users = &UserRepository{db: pool}
	driver.apiKeys = &APIKeyRepository{db: pool}
	driver.metars = &METARRepository{db: pool}

	return nil
}

// Users provides the PostgreSQL user repository implementation
func (driver *Driver) Users() user.Repository {
	return driver.users
}

// APIKeys provides the PostgreSQL API key repository implementation
func (driver *Driver) APIKeys() apikey.Repository {
	return driver.apiKeys
}

// METARs provides the PostgreSQL METAR repository implementation
func (driver *Driver) METARs() metar.Repository {
	return driver.metars
}

// Close discards the repository implementations and closes the database connection
func (driver *Driver) Close() {
	driver.users = nil
	driver.apiKeys = nil
	driver.metars = nil

	driver.db.Close()
	driver.db = nil
}

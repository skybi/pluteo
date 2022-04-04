package postgres

import (
	"context"
	"errors"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/skybi/data-server/internal/apikey"
	"github.com/skybi/data-server/internal/secret"
)

var keyLength = 64

// APIKeyRepository implements the apikey.Repository interface using PostgreSQL
type APIKeyRepository struct {
	db *pgxpool.Pool
}

var _ apikey.Repository = (*APIKeyRepository)(nil)

// Get retrieves multiple API keys
func (repo *APIKeyRepository) Get(ctx context.Context, offset, limit uint64) ([]*apikey.Key, uint64, error) {
	query := squirrel.Select("*").From("api_keys")
	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	} else if limit <= 0 {
		query = query.Limit(10)
	}
	sql, vals, err := query.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return nil, 0, err
	}

	var n uint64
	if err := repo.db.QueryRow(ctx, "select count(*) from api_keys").Scan(&n); err != nil {
		return nil, 0, err
	}
	if n == 0 {
		return []*apikey.Key{}, 0, nil
	}

	rows, err := repo.db.Query(ctx, sql, vals...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []*apikey.Key{}, n, nil
		}
		return nil, 0, err
	}

	keys := []*apikey.Key{}
	for rows.Next() {
		key, err := repo.rowToAPIKey(rows)
		if err != nil {
			return nil, 0, err
		}
		keys = append(keys, key)
	}

	return keys, n, nil
}

// GetByUserID retrieves multiple API keys of a specific user
func (repo *APIKeyRepository) GetByUserID(ctx context.Context, userID string, offset, limit uint64) ([]*apikey.Key, uint64, error) {
	query := squirrel.Select("*").From("api_keys").Where(squirrel.Eq{"user_id": userID})
	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	} else if limit <= 0 {
		query = query.Limit(10)
	}
	sql, vals, err := query.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return nil, 0, err
	}

	var n uint64
	if err := repo.db.QueryRow(ctx, "select count(*) from api_keys where user_id = $1", userID).Scan(&n); err != nil {
		return nil, 0, err
	}
	if n == 0 {
		return []*apikey.Key{}, 0, nil
	}

	rows, err := repo.db.Query(ctx, sql, vals...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []*apikey.Key{}, n, nil
		}
		return nil, 0, err
	}

	keys := []*apikey.Key{}
	for rows.Next() {
		key, err := repo.rowToAPIKey(rows)
		if err != nil {
			return nil, 0, err
		}
		keys = append(keys, key)
	}

	return keys, n, nil
}

// GetByID retrieves an API key by its ID
func (repo *APIKeyRepository) GetByID(ctx context.Context, id uuid.UUID) (*apikey.Key, error) {
	row := repo.db.QueryRow(ctx, "select * from api_keys where key_id = $1", id)
	key, err := repo.rowToAPIKey(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return key, nil
}

// GetByRawKey retrieves an API key by the raw bearer token
func (repo *APIKeyRepository) GetByRawKey(ctx context.Context, key string) (*apikey.Key, error) {
	hash, err := secret.Hash(key)
	if err != nil {
		// The raw key is no valid base64 string. This has the same effect as an invalid key.
		return nil, nil
	}

	row := repo.db.QueryRow(ctx, "select * from api_keys where api_key = $1", hash[:])
	obj, err := repo.rowToAPIKey(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return obj, nil
}

// Create creates a new API key
func (repo *APIKeyRepository) Create(ctx context.Context, create *apikey.Create) (*apikey.Key, string, error) {
	id := uuid.New()
	key, keyHash := secret.MustNew(keyLength)

	query := `
		insert into api_keys (key_id, api_key, user_id, description, quota, used_quota, rate_limit, capabilities)
		values ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := repo.db.Exec(
		ctx,
		query,
		id,
		keyHash[:],
		create.UserID,
		create.Description,
		create.Quota,
		0,
		create.RateLimit,
		create.Capabilities,
	)
	if err != nil {
		return nil, "", err
	}

	return &apikey.Key{
		ID:           id,
		Key:          keyHash[:],
		UserID:       create.UserID,
		Description:  create.Description,
		Quota:        create.Quota,
		UsedQuota:    0,
		RateLimit:    create.RateLimit,
		Capabilities: create.Capabilities,
	}, key, nil
}

// Update updates an API key
func (repo *APIKeyRepository) Update(ctx context.Context, id uuid.UUID, update *apikey.Update) (*apikey.Key, error) {
	// Simply re-fetch the API key if nothing should be changed
	if update.Description == nil && update.Quota == nil && update.UsedQuota == nil && update.RateLimit == nil && update.Capabilities == nil {
		return repo.GetByID(ctx, id)
	}

	// Build the SQL query
	query := squirrel.Update("api_keys").Where(squirrel.Eq{"key_id": id})
	if update.Description != nil {
		query = query.Set("description", *update.Description)
	}
	if update.Quota != nil {
		query = query.Set("quota", *update.Quota)
	}
	if update.UsedQuota != nil {
		query = query.Set("used_quota", *update.UsedQuota)
	}
	if update.RateLimit != nil {
		query = query.Set("rate_limit", *update.RateLimit)
	}
	if update.Capabilities != nil {
		query = query.Set("capabilities", *update.Capabilities)
	}
	sql, values, err := query.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return nil, err
	}

	// Perform the SQL query
	_, err = repo.db.Exec(ctx, sql, values...)
	if err != nil {
		return nil, err
	}

	// Re-fetch the API key
	return repo.GetByID(ctx, id)
}

// Delete deletes an API key by its ID
func (repo *APIKeyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := repo.db.Exec(ctx, "delete from api_keys where key_id = $1", id)
	return err
}

func (repo *APIKeyRepository) rowToAPIKey(row pgx.Row) (*apikey.Key, error) {
	obj := new(apikey.Key)
	if err := row.Scan(&obj.ID, &obj.Key, &obj.UserID, &obj.Description, &obj.Quota, &obj.UsedQuota, &obj.RateLimit, &obj.Capabilities); err != nil {
		return nil, err
	}
	return obj, nil
}

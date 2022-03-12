package postgres

import (
	"context"
	"errors"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/skybi/data-server/internal/user"
)

var (
	ErrMissingAPIKeyPolicy = errors.New("an initial API key policy is required")
)

// UserRepository implements the user.Repository for PostgreSQL
type UserRepository struct {
	db *pgxpool.Pool
}

var _ user.Repository = (*UserRepository)(nil)

// GetByID retrieves a user by their ID
func (repo *UserRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
	// Retrieve the user row itself
	userRow := repo.db.QueryRow(ctx, "select * from users where user_id = $1", id)
	userObj, err := repo.rowToUser(userRow)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	// Retrieve the corresponding API key policy and add it to the user object
	apiKeyPolicyRow := repo.db.QueryRow(ctx, "select * from user_api_key_policies where user_id = $1", id)
	apiKeyPolicyObj, err := repo.rowToAPIKeyPolicy(apiKeyPolicyRow)
	if err != nil {
		return nil, err
	}
	userObj.APIKeyPolicy = apiKeyPolicyObj

	return userObj, nil
}

// Create creates a new user
func (repo *UserRepository) Create(ctx context.Context, create *user.Create) (*user.User, error) {
	// Ensure an initial API key policy is provided
	if create.APIKeyPolicy == nil {
		return nil, ErrMissingAPIKeyPolicy
	}

	// Begin a new transaction
	tx, err := repo.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Create the user row itself
	createUserQuery := `
		insert into users (user_id, display_name, restricted)
		values ($1, $2, $3)
	`
	_, err = tx.Exec(ctx, createUserQuery, create.ID, create.DisplayName, false)
	if err != nil {
		return nil, err
	}

	// Create the corresponding API key policy row
	createAPIKeyPolicyQuery := `
		insert into user_api_key_policies (user_id, max_quota, max_rate_limit, allowed_capabilities)
		values ($1, $2, $3, $4)
	`
	_, err = tx.Exec(
		ctx,
		createAPIKeyPolicyQuery,
		create.ID,
		create.APIKeyPolicy.MaxQuota,
		create.APIKeyPolicy.MaxRateLimit,
		create.APIKeyPolicy.AllowedCapabilities,
	)
	if err != nil {
		return nil, err
	}

	// Commit the changes
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	cpy := *create.APIKeyPolicy
	return &user.User{
		ID:           create.ID,
		DisplayName:  create.DisplayName,
		APIKeyPolicy: &cpy,
		Restricted:   false,
	}, nil
}

// Update updates an existing user
func (repo *UserRepository) Update(ctx context.Context, id string, update *user.Update) (*user.User, error) {
	// Begin a new transaction
	tx, err := repo.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Update the user object itself if needed
	if update.DisplayName != nil || update.Restricted != nil {
		query := squirrel.Update("users").Where(squirrel.Eq{"user_id": id})
		if update.DisplayName != nil {
			query = query.Set("display_name", *update.DisplayName)
		}
		if update.Restricted != nil {
			query = query.Set("restricted", *update.Restricted)
		}

		sql, values, err := query.PlaceholderFormat(squirrel.Dollar).ToSql()
		if err != nil {
			return nil, err
		}
		_, err = tx.Exec(ctx, sql, values...)
		if err != nil {
			return nil, err
		}
	}

	// Update the users API key policy if needed
	if update.APIKeyPolicy != nil && (update.APIKeyPolicy.MaxQuota != nil || update.APIKeyPolicy.MaxRateLimit != nil || update.APIKeyPolicy.AllowedCapabilities != nil) {
		query := squirrel.Update("user_api_key_policies").Where(squirrel.Eq{"user_id": id})
		if update.APIKeyPolicy.MaxQuota != nil {
			query = query.Set("max_quota", *update.APIKeyPolicy.MaxQuota)
		}
		if update.APIKeyPolicy.MaxRateLimit != nil {
			query = query.Set("max_rate_limit", *update.APIKeyPolicy.MaxRateLimit)
		}
		if update.APIKeyPolicy.AllowedCapabilities != nil {
			query = query.Set("allowed_capabilities", *update.APIKeyPolicy.AllowedCapabilities)
		}

		sql, values, err := query.PlaceholderFormat(squirrel.Dollar).ToSql()
		if err != nil {
			return nil, err
		}
		_, err = tx.Exec(ctx, sql, values...)
		if err != nil {
			return nil, err
		}
	}

	// Commit the changes
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	// Re-fetch the user
	return repo.GetByID(ctx, id)
}

// Delete deletes a user by their ID
func (repo *UserRepository) Delete(ctx context.Context, id string) error {
	_, err := repo.db.Exec(ctx, "delete from users where user_id = $1", id)
	return err
}

func (repo *UserRepository) rowToUser(row pgx.Row) (*user.User, error) {
	obj := new(user.User)
	if err := row.Scan(&obj.ID, &obj.DisplayName, &obj.Restricted); err != nil {
		return nil, err
	}
	return obj, nil
}

func (repo *UserRepository) rowToAPIKeyPolicy(row pgx.Row) (*user.APIKeyPolicy, error) {
	obj := new(user.APIKeyPolicy)
	if err := row.Scan(nil, &obj.MaxQuota, &obj.MaxRateLimit, &obj.AllowedCapabilities); err != nil {
		return nil, err
	}
	return obj, nil
}
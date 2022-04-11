package postgres

import (
	"context"
	"errors"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/skybi/pluteo/internal/user"
)

var (
	ErrMissingAPIKeyPolicy = errors.New("an initial API key policy is required")
)

// UserRepository implements the user.Repository interface using PostgreSQL
type UserRepository struct {
	db *pgxpool.Pool
}

var _ user.Repository = (*UserRepository)(nil)

// Get retrieves multiple users
func (repo *UserRepository) Get(ctx context.Context, offset, limit uint64) ([]*user.User, uint64, error) {
	query := squirrel.Select(
		"users.user_id",
		"users.display_name",
		"users.restricted",
		"users.admin",
		"user_api_key_policies.max_quota",
		"user_api_key_policies.max_rate_limit",
		"user_api_key_policies.allowed_capabilities",
	).From("users").JoinClause("INNER JOIN user_api_key_policies ON users.user_id = user_api_key_policies.user_id")
	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	} else if limit <= 0 {
		query = query.Limit(10)
	}
	sql, vals, err := query.ToSql()
	if err != nil {
		return nil, 0, err
	}

	var n uint64
	if err := repo.db.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&n); err != nil {
		return nil, 0, err
	}
	if n == 0 {
		return []*user.User{}, 0, nil
	}

	rows, err := repo.db.Query(ctx, sql, vals...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []*user.User{}, n, nil
		}
		return nil, 0, err
	}

	users := []*user.User{}
	for rows.Next() {
		obj := &user.User{
			APIKeyPolicy: &user.APIKeyPolicy{},
		}
		err = rows.Scan(
			&obj.ID,
			&obj.DisplayName,
			&obj.Restricted,
			&obj.Admin,
			&obj.APIKeyPolicy.MaxQuota,
			&obj.APIKeyPolicy.MaxRateLimit,
			&obj.APIKeyPolicy.AllowedCapabilities,
		)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, obj)
	}

	return users, n, nil
}

// GetByID retrieves a user by their ID
func (repo *UserRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
	// Retrieve the user row itself
	userRow := repo.db.QueryRow(ctx, "SELECT * FROM users WHERE user_id = $1", id)
	userObj, err := repo.rowToUser(userRow)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	// Retrieve the corresponding API key policy and add it to the user object
	apiKeyPolicyRow := repo.db.QueryRow(ctx, "SELECT * FROM user_api_key_policies WHERE user_id = $1", id)
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
	_, err = tx.Exec(ctx, "INSERT INTO users VALUES ($1, $2, $3, $4)", create.ID, create.DisplayName, false, create.Admin)
	if err != nil {
		return nil, err
	}

	// Create the corresponding API key policy row
	_, err = tx.Exec(
		ctx,
		"INSERT INTO user_api_key_policies VALUES ($1, $2, $3, $4)",
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
		Admin:        create.Admin,
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
	if update.DisplayName != nil || update.Restricted != nil || update.Admin != nil {
		query := squirrel.Update("users").Where(squirrel.Eq{"user_id": id})
		if update.DisplayName != nil {
			query = query.Set("display_name", *update.DisplayName)
		}
		if update.Restricted != nil {
			query = query.Set("restricted", *update.Restricted)
		}
		if update.Admin != nil {
			query = query.Set("admin", *update.Admin)
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
	_, err := repo.db.Exec(ctx, "DELETE FROM users WHERE user_id = $1", id)
	return err
}

func (repo *UserRepository) rowToUser(row pgx.Row) (*user.User, error) {
	obj := new(user.User)
	if err := row.Scan(&obj.ID, &obj.DisplayName, &obj.Restricted, &obj.Admin); err != nil {
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

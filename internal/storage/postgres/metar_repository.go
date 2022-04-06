package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/skybi/data-server/internal/metar"
)

// METARRepository implements the metar.Repository interface using PostgreSQL
type METARRepository struct {
	db *pgxpool.Pool
}

var _ metar.Repository = (*METARRepository)(nil)

// GetByFilter retrieves multiple METARs following a filter, ordered by their issuing date (descending).
// If limit <= 0, a default limit value of 10 is used.
func (repo *METARRepository) GetByFilter(ctx context.Context, filter *metar.Filter, limit uint64) ([]*metar.METAR, uint64, error) {
	// Construct the raw (unlimited) SQL query
	rawQuery := squirrel.Select("*").From("metars").OrderBy("issued_at DESC")
	if filter.StationID != nil {
		rawQuery = rawQuery.Where(squirrel.Eq{"station_id": *filter.StationID})
	}
	if filter.IssuedBefore != nil {
		rawQuery = rawQuery.Where(squirrel.Lt{"issued_at": *filter.IssuedBefore})
	}
	if filter.IssuedAfter != nil {
		rawQuery = rawQuery.Where(squirrel.Gt{"issued_at": *filter.IssuedAfter})
	}
	rawSQL, rawVals, err := rawQuery.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return nil, 0, err
	}

	// Construct the limited query
	limitedQuery := rawQuery
	if limit > 0 {
		limitedQuery = limitedQuery.Limit(limit)
	} else if limit <= 0 {
		limitedQuery = limitedQuery.Limit(10)
	}
	limitedSQL, limitedVals, err := limitedQuery.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return nil, 0, err
	}

	// Fetch the total amount of METARs that matches the given filter
	var n uint64
	if err := repo.db.QueryRow(ctx, rawSQL, rawVals...).Scan(&n); err != nil {
		return nil, 0, err
	}
	if n == 0 {
		return []*metar.METAR{}, 0, nil
	}

	// Fetch the METAR objects themselves
	rows, err := repo.db.Query(ctx, limitedSQL, limitedVals...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []*metar.METAR{}, n, nil
		}
		return nil, 0, err
	}
	objs := []*metar.METAR{}
	for rows.Next() {
		obj, err := repo.rowToMETAR(rows)
		if err != nil {
			return nil, 0, err
		}
		objs = append(objs, obj)
	}

	return objs, n, nil
}

// GetByID retrieves a METAR by its ID
func (repo *METARRepository) GetByID(ctx context.Context, id uuid.UUID) (*metar.METAR, error) {
	row := repo.db.QueryRow(ctx, "SELECT * FROM metars WHERE metar_id = $1", id)
	obj, err := repo.rowToMETAR(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return obj, nil
}

// Create creates new METARs based on their raw text representation.
// All raw strings are sanitized (leading and trailing spaces are trimmed).
// This method also returns the indexes of the METARs that already exist in the database and thus were not inserted.
func (repo *METARRepository) Create(ctx context.Context, raw []string) ([]*metar.METAR, []uint, error) {
	txn, err := repo.db.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer txn.Rollback(ctx)

	metars := make([]*metar.METAR, 0, len(raw))
	uniqueViolations := []uint{}

	for i, str := range raw {
		// Parse the raw string into a metar.METAR object
		obj, err := metar.OfString(str)
		if err != nil {
			var formatErr *metar.FormatError
			if errors.As(err, &formatErr) {
				return nil, nil, &metar.FormatError{Wrapping: fmt.Errorf("error in METAR no. %d: %s", i, err.Error())}
			}
			return nil, nil, err
		}

		// Insert the METAR into the database
		tag, err := txn.Exec(ctx, "INSERT INTO metars VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING", obj.ID, obj.StationID, obj.IssuedAt, obj.Raw)
		if err != nil {
			return nil, nil, err
		}
		if tag.RowsAffected() == 0 {
			uniqueViolations = append(uniqueViolations, uint(i))
			continue
		}

		metars = append(metars, obj)
	}

	if err := txn.Commit(ctx); err != nil {
		return nil, nil, err
	}

	return metars, uniqueViolations, nil
}

// Delete deletes a METAR by its ID
func (repo *METARRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := repo.db.Exec(ctx, "DELETE FROM metars WHERE metar_id = $1", id)
	return err
}

func (repo *METARRepository) rowToMETAR(row pgx.Row) (*metar.METAR, error) {
	obj := new(metar.METAR)
	if err := row.Scan(&obj.ID, &obj.StationID, &obj.IssuedAt, &obj.Raw); err != nil {
		return nil, err
	}
	return obj, nil
}

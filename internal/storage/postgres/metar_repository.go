package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/skybi/pluteo/internal/metar"
)

// METARRepository implements the metar.Repository interface using PostgreSQL
type METARRepository struct {
	db *pgxpool.Pool
}

var _ metar.Repository = (*METARRepository)(nil)

// GetByFilter retrieves multiple METARs following a filter, ordered by their issuing date (descending).
// If limit <= 0, a default limit value of 10 is used.
func (repo *METARRepository) GetByFilter(ctx context.Context, filter *metar.Filter, limit uint64) ([]*metar.METAR, uint64, error) {
	// Construct the SQL queries
	countQuery := squirrel.Select("COUNT(*)").From("metars")
	query := squirrel.Select("*").From("metars").OrderBy("issued_at DESC")
	if filter.StationID != nil {
		countQuery = countQuery.Where(squirrel.Eq{"station_id": *filter.StationID})
		query = query.Where(squirrel.Eq{"station_id": *filter.StationID})
	}
	if filter.IssuedBefore != nil {
		countQuery = countQuery.Where(squirrel.Lt{"issued_at": *filter.IssuedBefore})
		query = query.Where(squirrel.Lt{"issued_at": *filter.IssuedBefore})
	}
	if filter.IssuedAfter != nil {
		countQuery = countQuery.Where(squirrel.Gt{"issued_at": *filter.IssuedAfter})
		query = query.Where(squirrel.Gt{"issued_at": *filter.IssuedAfter})
	}
	if limit > 0 {
		query = query.Limit(limit)
	} else if limit <= 0 {
		query = query.Limit(10)
	}
	countSQL, countVals, err := countQuery.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return nil, 0, err
	}
	sql, vals, err := query.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return nil, 0, err
	}

	// Fetch the total amount of METARs that matches the given filter
	var n uint64
	if err := repo.db.QueryRow(ctx, countSQL, countVals...).Scan(&n); err != nil {
		return nil, 0, err
	}
	if n == 0 {
		return []*metar.METAR{}, 0, nil
	}

	// Fetch the METAR objects themselves
	rows, err := repo.db.Query(ctx, sql, vals...)
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
				return nil, nil, &metar.FormatError{
					Wrapping: fmt.Errorf("error in METAR no. %d: %s", i, err.Error()),
					Index:    i,
				}
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

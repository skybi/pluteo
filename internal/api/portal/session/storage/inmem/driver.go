package inmem

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"github.com/hashicorp/go-memdb"
	"github.com/skybi/data-server/internal/api/portal/session"
	"github.com/skybi/data-server/internal/random"
	"time"
)

var tokenLength = 64

var dbSchema = &memdb.DBSchema{
	Tables: map[string]*memdb.TableSchema{
		"sessions": {
			Name: "sessions",
			Indexes: map[string]*memdb.IndexSchema{
				"id": {
					Name:         "id",
					Unique:       true,
					AllowMissing: false,
					Indexer:      &memdb.StringFieldIndex{Field: "Token"},
				},
				"sessionID": {
					Name:         "sessionID",
					Unique:       true,
					AllowMissing: true,
					Indexer:      &memdb.StringFieldIndex{Field: "SessionID"},
				},
				"userID": {
					Name:         "userID",
					Unique:       false,
					AllowMissing: false,
					Indexer:      &memdb.StringFieldIndex{Field: "UserID"},
				},
				"expires": {
					Name:         "expires",
					Unique:       false,
					AllowMissing: false,
					Indexer:      &memdb.IntFieldIndex{Field: "Expires"},
				},
			},
		},
	},
}

// Driver represents the in-memory session storage driver built using hashicorp/go-memdb
type Driver struct {
	db *memdb.MemDB
}

var _ session.Storage = (*Driver)(nil)

// New creates a new empty in-memory session storage driver
func New() (*Driver, error) {
	db, err := memdb.NewMemDB(dbSchema)
	if err != nil {
		return nil, err
	}
	return &Driver{db}, nil
}

// GetByRawToken retrieves a session by its raw (prior hashing) token
func (driver *Driver) GetByRawToken(_ context.Context, rawToken string) (*session.Session, error) {
	hash := hashToken(rawToken)

	txn := driver.db.Txn(false)
	obj, err := txn.First("sessions", "id", hash)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, nil
	}

	return obj.(*session.Session), nil
}

// Create creates a new session
func (driver *Driver) Create(_ context.Context, userID, sessionID string, expires int64) (string, error) {
	rawToken := random.String(tokenLength, random.CharsetTokens)
	token := hashToken(rawToken)

	ses := &session.Session{
		Token:     token,
		SessionID: sessionID,
		UserID:    userID,
		Expires:   expires,
	}

	txn := driver.db.Txn(true)
	defer txn.Abort()
	if err := txn.Insert("sessions", ses); err != nil {
		return "", err
	}
	txn.Commit()

	return rawToken, nil
}

// TerminateBySessionID terminates a session by its session ID
func (driver *Driver) TerminateBySessionID(_ context.Context, sessionID string) error {
	txn := driver.db.Txn(true)
	defer txn.Abort()
	if _, err := txn.DeleteAll("sessions", "sessionID", sessionID); err != nil {
		return err
	}
	txn.Commit()
	return nil
}

// TerminateByUserID terminates all sessions of a specific user ID
func (driver *Driver) TerminateByUserID(_ context.Context, userID string) error {
	txn := driver.db.Txn(true)
	defer txn.Abort()
	if _, err := txn.DeleteAll("sessions", "userID", userID); err != nil {
		return err
	}
	txn.Commit()
	return nil
}

// TerminateExpired terminates all sessions that are expired
func (driver *Driver) TerminateExpired(_ context.Context) (int, error) {
	txn := driver.db.Txn(true)
	defer txn.Abort()

	it, err := txn.LowerBound("sessions", "expires", 0)
	if err != nil {
		return 0, err
	}

	now := time.Now().Unix()
	deleted := 0
	for obj := it.Next(); obj != nil; obj = it.Next() {
		ses := obj.(*session.Session)
		if ses.Expires > now {
			break
		}
		if err := txn.Delete("sessions", ses); err != nil {
			return 0, err
		}
		deleted++
	}

	txn.Commit()
	return deleted, nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

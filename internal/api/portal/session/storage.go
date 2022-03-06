package session

import "context"

// Storage defines the session storage API
type Storage interface {
	// GetByRawToken retrieves a session by its raw (prior hashing) token
	GetByRawToken(ctx context.Context, rawToken string) (*Session, error)

	// Create creates a new session
	Create(ctx context.Context, userID, sessionID string, expires int64) (string, error)

	// TerminateBySessionID terminates a session by its session ID
	TerminateBySessionID(ctx context.Context, sessionID string) error

	// TerminateByUserID terminates all sessions of a specific user ID
	TerminateByUserID(ctx context.Context, userID string) error

	// TerminateExpired terminates all sessions that are expired
	TerminateExpired(ctx context.Context) (int, error)
}

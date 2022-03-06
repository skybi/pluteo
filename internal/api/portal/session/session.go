package session

// Session represents a user session at the portal API.
// A session is identified by its token rather than its session ID as the session ID is an optional field whose presence
// depends on whether the OIDC provider implements OpenID session management.
type Session struct {
	Token     string
	SessionID string
	UserID    string
	Expires   int64
}

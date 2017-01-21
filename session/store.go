package session

import (
	"net/http"
)

// Store is an interface for custom session stores.
// See CookieStore and FilesystemStore for examples.
type Store interface {
	// Get should return a cached session.
	Get(r *http.Request, name string) (*RawSession, error)

	// New should create and return a new session.
	// Note that New should never return a nil session, even in the case of
	// an error if using the Registry infrastructure to cache the session.
	New(r *http.Request, name string) (*RawSession, error)

	// Save should persist session to the underlying store implementation.
	Save(r *http.Request, w http.ResponseWriter, s *RawSession) error
}

package session

import (
	"net/http"
	"github.com/gorilla/securecookie"
	"fmt"
)

// CookieStore stores sessions using secure cookies.
type CookieStore struct {
	Codecs  []securecookie.Codec
	*Options // default configuration
}

// NewCookieStore returns a new CookieStore.
//
// Keys are defined in pairs to allow key rotation, but the common case is
// to set a single authentication key and optionally an encryption key.
//
// The first key in a pair is used for authentication and the second for
// encryption. The encryption key can be set to nil or omitted in the last
// pair, but the authentication key is required in all pairs.
//
// It is recommended to use an authentication key with 32 or 64 bytes.
// The encryption key, if set, must be either 16, 24, or 32 bytes to select
// AES-128, AES-192, or AES-256 modes.
//
// Use the convenience function securecookie.GenerateRandomKey() to create
// strong keys.
func NewCookieStore(keyPairs ...[]byte) *CookieStore {
	cs := &CookieStore{
		Codecs: securecookie.CodecsFromPairs(keyPairs...),
		Options: &Options{
			Path:   "/",
			MaxAge: sessionExpire,
		},
	}

	cs.MaxAge(cs.Options.MaxAge)
	return cs
}

// Get returns a session for the given name.
//
// It returns a new session if the sessions doesn't exist. Access IsNew on
// the session to check if it is an existing session or a new one.
//
// It returns a new session and an error if the session exists but could
// not be decoded.
func (store *CookieStore) Get(r *http.Request, name string, s *session) error {
	// Copy options.
	options := *store.Options
	s.Options = &options

	cookie, err := r.Cookie(name)
	if err != nil {
		return err
	}
	// Decode to get contents.
	err = securecookie.DecodeMulti(name, cookie.Value, &s.Contents, store.Codecs...)
	return err
}

// Save adds a single session to the response.
func (store *CookieStore) Save(r *http.Request, w http.ResponseWriter, s *session) error {
	// Encode to put contents.
	value, err := securecookie.EncodeMulti(s.Name, s.Contents, store.Codecs...)
	if err != nil {
		return err
	}
	http.SetCookie(w, newCookie(s.Name, value, s.Options))
	return nil
}

// MaxAge sets the maximum age for cookie.
// Individual sessions can be deleted by setting Options.MaxAge
// = -1 for that session.
func (store *CookieStore) MaxAge(age int) {
	store.Options.MaxAge = age

	// Set the maxAge for each securecookie instance.
	for _, codec := range store.Codecs {
		if c, ok := codec.(*securecookie.SecureCookie); ok {
			c.MaxAge(age)
		} else {
			fmt.Printf("Can't change MaxAge on codec %v\n", codec)
		}
	}
}

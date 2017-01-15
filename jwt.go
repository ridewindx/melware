package melware

import (
	"github.com/ridewindx/mel"
	"gopkg.in/dgrijalva/jwt-go.v3"
	"net/http"
	"strings"
	"time"
)

// JWT provides a Json-Web-Token authentication implementation.
// On failure, a 401 HTTP response is returned.
// On success, the wrapped middleware is called, and the userID is made available as
// c.Get("userID").(string).
// Users can get a token by posting a json request to LoginHandler. The token then needs to be passed in
// the Authentication header.
type JWT struct {
	// Realm specifies the realm name to display to the user.
	// Required.
	Realm string

	// SigningAlgorithm specifies signing algorithm.
	// Optional. Default is HS256.
	SigningAlgorithm string

	// Key specifies the secret key used for signing.
	// Required.
	Key []byte

	// Timeout specifies the duration that a token is valid.
	// Optional. Defaults to one hour.
	Timeout time.Duration

	// MaxRefresh specifies the maximum duration in which the client can refresh its token.
	// This means that the maximum validity timespan for a token is MaxRefresh + Timeout.
	// Optional. Defaults to 0 meaning not refreshable.
	MaxRefresh time.Duration

	// Authenticator specifies the callback that should perform the authentication
	// of the user based on userID and password.
	// Must return true on success, false on failure.
	// Required. Option return user id, if so, user id will be stored in Claim Array.
	Authenticator func(userID string, password string, c *mel.Context) (string, bool)
}

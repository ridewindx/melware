package melware

import (
	"github.com/ridewindx/mel"
	"gopkg.in/dgrijalva/jwt-go.v3"
	"net/http"
	"strings"
	"time"
	"github.com/gin-gonic/gin"
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

	// Callback function that should perform the authorization of the authenticated user. Called
	// only after an authentication success. Must return true on success, false on failure.
	// Optional, default to success.
	Authorizator func(userID string, c *gin.Context) bool

	// Callback function that will be called during login.
	// Using this function it is possible to add additional payload data to the webtoken.
	// The data is then made available during requests via c.Get("JWT_PAYLOAD").
	// Note that the payload is not encrypted.
	// The attributes mentioned on jwt.io can't be used as keys for the map.
	// Optional, by default no additional data will be set.
	PayloadFunc func(userID string) map[string]interface{}

	// User can define own Unauthorized func.
	Unauthorized func(*gin.Context, int, string)

	// TokenLookup is a string in the form of "<source>:<name>" that is used
	// to extract token from the request.
	// Optional. Default value "header:Authorization".
	// Possible values:
	// - "header:<name>"
	// - "query:<name>"
	// - "cookie:<name>"
	TokenLookup string
}

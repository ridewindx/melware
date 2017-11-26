package melware

import (
	"github.com/ridewindx/mel"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"strings"
	"time"
	"errors"
)

// JWT provides a Json-Web-Token authentication implementation.
// On failure, a 401 HTTP response is returned.
// On success, the wrapped middleware is called, and the userID is made available as
// c.Get("userID").(string).
// Users can get a token by posting a json request to LoginHandler. The token then needs to be passed in
// the Authentication header.
type JWT struct {
	// Realm specifies the realm name to display to the user.
	// Optional.
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

	// Authenticate specifies the callback that should perform the authentication
	// of the user based on request context.
	// Must return nil error on success, error on failure.
	// Required. Optional return user id, if so, user id will be stored in Claim Array.
	Authenticate func(c *mel.Context) (string, error)

	// Authorize specifies the callback that should perform the authorization
	// of the authenticated user.
	// Must return true on success, false on failure.
	// Optional. Default to always return success.
	Authorize func(userID string, c *mel.Context) bool

	// PayloadFunc specifies the callback that will be called during login.
	// It is useful for adding additional payload data to the token.
	// The data is then made available during requests via ExtractClaims(PayloadKey).
	// Note that the payload is not encrypted.
	// Optional. By default no additional payload will be added.
	PayloadFunc func(userID string) map[string]interface{}

	// Unauthorized specifies the unauthorized function.
	Unauthorized func(*mel.Context, int, string)

	// TokenBearer is a string in the form of "<source>:<name>"
	// that is used to extract token from the request.
	// Optional. Default to "header:Authorization".
	// Possible values:
	// - "header:<name>"
	// - "query:<name>"
	// - "cookie:<name>"
	TokenBearer string

	// PayloadKey specifies the key when puts JWT payload into Context.
	// Optional. Default to "JWT_PAYLOAD".
	PayloadKey string

	extractToken func(*mel.Context) (string, error)
}

func (j *JWT) init() {
	if j.SigningAlgorithm == "" {
		j.SigningAlgorithm = "HS256"
	}

	if j.Key == nil {
		panic("Secret key is required")
	}

	if j.Timeout == 0 {
		j.Timeout = time.Hour
	}

	if j.Authenticate == nil {
		panic("Authenticate funciton is required")
	}

	if j.Authorize == nil {
		j.Authorize = func(userId string, c *mel.Context) bool {
			return true
		}
	}

	if j.Unauthorized == nil {
		j.Unauthorized = func(c *mel.Context, code int, message string) {
			c.JSON(code, mel.Map{
				"code":    code,
				"message": message,
			})
		}
	}

	if j.TokenBearer == "" {
		j.TokenBearer = "header:Authorization"
	}

	parts := strings.Split(j.TokenBearer, ":")
	key := parts[1]
	switch parts[0] {
	case "header":
		j.extractToken = func(c *mel.Context) (string, error) {
			authHeader := c.Request.Header.Get(key)

			if len(authHeader) == 0 {
				return "", errors.New("empty auth header")
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if !(len(parts) == 2 && parts[0] == "Bearer") {
				return "", errors.New("invalid auth header")
			}

			return parts[1], nil
		}

	case "query":
		j.extractToken = func(c *mel.Context) (string, error) {
			token := c.Query(key)

			if len(token) == 0 {
				return "", errors.New("empty query token")
			}

			return token, nil
		}

	case "cookie":
		j.extractToken = func(c *mel.Context) (string, error) {
			cookie, _ := c.Cookie(key)

			if len(cookie) == 0 {
				return "", errors.New("empty cookie token")
			}

			return cookie, nil
		}

	default:
		panic("Invalid token source")
	}

	if len(j.PayloadKey) == 0 {
		j.PayloadKey = "JWT_PAYLOAD"
	}
}

// Middleware returns a middleware that authorizes tokens.
func (j *JWT) Middleware() mel.Handler {
	j.init()

	return func(c *mel.Context) {
		token, err := j.parseToken(c)

		if err != nil {
			j.unauthorized(c, http.StatusUnauthorized, err.Error())
			return
		}

		claims := token.Claims.(jwt.MapClaims)

		userId := claims["id"].(string)
		c.Set(j.PayloadKey, claims)
		c.Set("userID", userId)

		if !j.Authorize(userId, c) {
			j.unauthorized(c, http.StatusForbidden, "You don't have permission to access.")
			return
		}

		c.Next()
		return
	}
}

// LoginHandler can be used by clients to get a jwt token.
// Reply will be of the form {"token": "TOKEN"}.
func (j *JWT) LoginHandler() mel.Handler {
	j.init()

	return func(c *mel.Context) {
		userID, err := j.Authenticate(c)
		if err != nil {
			j.unauthorized(c, http.StatusUnauthorized, err.Error())
			return
		}

		// Create the token
		token := jwt.New(jwt.GetSigningMethod(j.SigningAlgorithm))
		claims := token.Claims.(jwt.MapClaims)

		if j.PayloadFunc != nil {
			for key, value := range j.PayloadFunc(userID) {
				claims[key] = value
			}
		}

		expire := time.Now().Add(j.Timeout)
		claims["id"] = userID
		claims["exp"] = expire.Unix()
		claims["iat"] = time.Now().Unix()

		tokenStr, err := token.SignedString(j.Key)

		if err != nil {
			j.unauthorized(c, http.StatusUnauthorized, "Create JWT token faild")
			return
		}

		c.JSON(http.StatusOK, mel.Map{
			"token":      tokenStr,
			"expires_at": expire.Format(time.RFC3339),
		})
	}
}

// RefreshHandler can be used to refresh a token.
// The token still needs to be valid on refresh.
// Shall be put under an endpoint that is using the Middleware.
// Reply will be of the form {"token": "TOKEN"}.
func (j *JWT) RefreshHandler(c *mel.Context) {
	token, _ := j.parseToken(c)
	claims := token.Claims.(jwt.MapClaims)

	iat := int64(claims["iat"].(int64))

	if iat < time.Now().Add(-j.MaxRefresh).Unix() {
		j.unauthorized(c, http.StatusUnauthorized, "Token is expired")
		return
	}

	// Refresh expiration time
	expire := time.Now().Add(j.Timeout)
	claims["exp"] = expire.Unix()

	// Create the token
	alg := jwt.GetSigningMethod(j.SigningAlgorithm)
	newToken := jwt.NewWithClaims(alg, claims)

	tokenStr, err := newToken.SignedString(j.Key)
	if err != nil {
		j.unauthorized(c, http.StatusUnauthorized, "Create JWT Token faild")
		return
	}

	c.JSON(http.StatusOK, mel.Map{
		"token":      tokenStr,
		"expires_at": expire.Format(time.RFC3339),
	})
}

// ExtractClaims extracts the JWT claims.
func (j *JWT) ExtractClaims(c *mel.Context) jwt.MapClaims {
	claims, ok := c.Get(j.PayloadKey)
	if ok {
		return claims.(jwt.MapClaims)
	} else {
		return nil
	}
}

func (j *JWT) parseToken(c *mel.Context) (*jwt.Token, error) {
	tokenStr, err := j.extractToken(c)

	if err != nil {
		return nil, err
	}

	return jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if jwt.GetSigningMethod(j.SigningAlgorithm) != token.Method {
			return nil, errors.New("invalid signing algorithm")
		}

		return j.Key, nil
	})
}

func (mw *JWT) unauthorized(c *mel.Context, code int, message string) {
	c.Header("WWW-Authenticate", "JWT realm="+mw.Realm)
	c.Abort()

	mw.Unauthorized(c, code, message)

	return
}

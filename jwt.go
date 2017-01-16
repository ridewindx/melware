package melware

import (
	"github.com/ridewindx/mel"
	"gopkg.in/dgrijalva/jwt-go.v3"
	"net/http"
	"strings"
	"time"
	"errors"
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
	MaxRefresh   time.Duration

	// Authenticator specifies the callback that should perform the authentication
	// of the user based on userID and password.
	// Must return true on success, false on failure.
	// Required. Option return user id, if so, user id will be stored in Claim Array.
	Authenticate func(userID string, password string, c *mel.Context) (string, bool)

	// Authorizator specifies the callback that should perform the authorization
	// of the authenticated user.
	// Must return true on success, false on failure.
	// Optional. Default to always return success.
	Authorizator func(userID string, c *gin.Context) bool

	// PayloadFunc specifies the callback that will be called during login.
	// It is useful for adding additional payload data to the token.
	// The data is then made available during requests via c.Get("JWT_PAYLOAD").
	// Note that the payload is not encrypted.
	// Optional. By default no additional payload will be added.
	PayloadFunc func(userID string) map[string]interface{}

	// Unauthorized specifies the unauthorized function.
	Unauthorized func(*gin.Context, int, string)

	// TokenLookup is a string in the form of "<source>:<name>" that is used
	// to extract token from the request.
	// Optional. Default value "header:Authorization".
	// Possible values:
	// - "header:<name>"
	// - "query:<name>"
	// - "cookie:<name>"
	TokenLookup string

	extractToken func(*mel.Context) (string, error)
}

type Login struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

func (j *JWT) init() {
	parts := strings.Split(j.TokenLookup, ":")
	key := parts[1]
	switch parts[0] {
	case "header":
		j.extractToken = func(c *mel.Context) (string, error) {
			authHeader := c.Request.Header.Get(key)

			if len(authHeader) == 0 {
				return "", errors.New("Empty auth header")
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if !(len(parts) == 2 && parts[0] == "Bearer") {
				return "", errors.New("Invalid auth header")
			}

			return parts[1], nil
		}
	case "query":
		j.extractToken = func(c *mel.Context) (string, error) {
			token := c.Query(key)

			if len(token) == 0 {
				return "", errors.New("Empty query token")
			}

			return token, nil
		}
	case "cookie":
		j.extractToken = func(c *mel.Context) (string, error) {
			cookie, _ := c.Cookie(key)

			if len(cookie) == 0 {
				return "", errors.New("Empty cookie token")
			}

			return cookie, nil
		}
	}


	return nil
}

// Middleware returns a middleware that authorizes tokens.
func (j *JWT) Middleware() mel.Handler {
	j.init()

	return func(c *gin.Context) {
		token, err := j.parseToken(c)

		if err != nil {
			j.unauthorized(c, http.StatusUnauthorized, err.Error())
			return
		}

		claims := token.Claims.(jwt.MapClaims)

		userId := claims["id"].(string)
		c.Set("JWT_PAYLOAD", claims)
		c.Set("userID", userId)

		if !j.Authorizator(userId, c) {
			j.unauthorized(c, http.StatusForbidden, "You don't have permission to access.")
			return
		}

		c.Next()
		return
	}
}

// LoginHandler can be used by clients to get a jwt token.
// Payload needs to be json in the form of {"username": "USERNAME", "password": "PASSWORD"}.
// Reply will be of the form {"token": "TOKEN"}.
func (j *JWT) LoginHandler() mel.Handler {
	j.init()

	return func(c *mel.Context) {
		var login Login

		if c.BindJSON(&login) != nil {
			j.unauthorized(c, http.StatusBadRequest, "Missing username or password")
			return
		}

		userID, ok := j.Authenticate(login.Username, login.Password, c)
		if !ok {
			j.unauthorized(c, http.StatusUnauthorized, "Invalid Username / Password")
			return
		}

		// Create the token
		token := jwt.New(jwt.GetSigningMethod(j.SigningAlgorithm))
		claims := token.Claims.(jwt.MapClaims)

		if j.PayloadFunc != nil {
			for key, value := range j.PayloadFunc(login.Username) {
				claims[key] = value
			}
		}

		if len(userID) == 0 {
			userID = login.Username
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
			"token":  tokenStr,
			"expires_at": expire.Format(time.RFC3339),
		})
	}
}

func (j *JWT) parseToken(c *mel.Context) (*jwt.Token, error) {
	tokenStr, err := j.extractToken(c)

	if err != nil {
		return nil, err
	}

	return jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if jwt.GetSigningMethod(j.SigningAlgorithm) != token.Method {
			return nil, errors.New("Invalid signing algorithm")
		}

		return j.Key, nil
	})
}

func (mw *JWT) unauthorized(c *mel.Context, code int, message string) {
	if mw.Realm == "" {
		mw.Realm = "mel jwt"
	}

	c.Header("WWW-Authenticate", "JWT realm="+mw.Realm)
	c.Abort()

	mw.Unauthorized(c, code, message)

	return
}

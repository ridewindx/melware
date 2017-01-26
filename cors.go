package melware

import (
	"time"
	"strings"
	"net/http"
	"fmt"
	"github.com/ridewindx/mel"
)

type cors struct {
	// AllowedOrigins is a slice of origins that a cors request can be executed from.
	// An origin may contain a wildcard (*) to replace 0 or more characters
	// (e.g., http://*.domain.com). Only one wildcard can be used per origin.
	// Default value is ["*"], i.e., all origins are allowed.
	AllowOrigins []string

	// AllowOriginFunc is a custom function to validate the origin. It take the origin
	// as argument and returns true if allowed or false otherwise.
	// It has lower precedence than AllowOrigins.
	AllowOriginFunc func(origin string) bool

	// AllowedMethods is a slice of methods the client is allowed to use with
	// cross-domain requests.
	// Default to {"GET", "POST", "HEAD"}.
	AllowMethods []string

	// AllowedHeaders is slice of non simple headers the client is allowed to use with
	// cross-domain requests.
	// Default to {"Origin", "Accept", "Content-Type"}.
	AllowHeaders []string

	// AllowCredentials indicates whether the request can include user credentials like
	// cookies, HTTP authentication or client side SSL certificates.
	AllowCredentials bool

	// ExposedHeaders indicates which headers are safe to expose to the API of a CORS
	// API specification
	ExposeHeaders []string

	// MaxAge indicates how long (in seconds) the results of a preflight request
	// can be cached
	MaxAge time.Duration

	allowAllOrigins bool
	normalHeaders    http.Header
	preflightHeaders http.Header
}

func Cors() *cors {
	return &cors{
		AllowOrigins: {"*"},
		AllowMethods: {"GET", "POST", "HEAD"},
		AllowHeaders: {"Origin", "Accept", "Content-Type"},
		AllowCredentials: false,
		MaxAge: 12 * time.Hour,
	}
}

func (c *cors) Middleware() mel.Handler {
	c.AllowOrigins = c.normalizeStrs(c.AllowOrigins)
	if len(c.AllowOrigins) == 1 && c.AllowOrigins[0] == "*" {
		c.allowAllOrigins = true
		if c.AllowOriginFunc != nil {
			panic("All origins are allowed, no predicate function needed")
		}
	} else if len(c.AllowOrigins) > 0 {
		for _, origin := range c.AllowOrigins {
			if origin == "*" {
				panic("All origins for cors are allowed, no individual origins needed")
			} else if !strings.HasPrefix(origin, "http://") && !strings.HasPrefix(origin, "https://") {
				panic("Origin must have prefix 'http://' or 'https://'")
			}
		}
	} else if c.AllowOriginFunc == nil {
		panic("No origin is allowed")
	}

	c.normalHeaders = c.generateNormalHeaders()
	c.preflightHeaders = c.generatePreflightHeaders()

	return func(ctx *mel.Context) {
		origin := ctx.Request.Header.Get("Origin")
		if len(origin) == 0 { // request is not a CORS request
			return
		}
		if !c.validateOrigin(origin) {
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}

		if ctx.Request.Method == "OPTIONS" {
			for key, value := range c.preflightHeaders {
				ctx.Header(key, value)
			}
			defer ctx.AbortWithStatus(http.StatusOK)
		} else {
			for key, value := range c.normalHeaders {
				ctx.Header(key, value)
			}
		}

		if !c.allowAllOrigins && !c.AllowCredentials {
			ctx.Header("Access-Control-Allow-Origin", origin)
		}

		ctx.Next()
	}
}

func (c *cors) validateOrigin(origin string) bool {
	if c.allowAllOrigins {
		return true
	}
	for _, value := range c.AllowOrigins {
		if value == origin {
			return true
		}
	}
	if c.AllowOriginFunc != nil {
		return c.AllowOriginFunc(origin)
	}
	return false
}

func (c *cors) normalizeStrs(strs []string) []string {
	if strs == nil {
		return nil
	}
	set := make(map[string]bool)
	var normalized []string
	for _, str := range strs {
		str = strings.TrimSpace(str)
		str = strings.ToLower(str)
		if _, seen := set[str]; !seen {
			normalized = append(normalized, str)
			set[str] = true
		}
	}
	return normalized
}

func (c *cors) generateNormalHeaders() http.Header {
	headers := make(http.Header)
	if c.AllowCredentials {
		headers.Set("Access-Control-Allow-Credentials", "true")
	}
	if len(c.ExposeHeaders) > 0 {
		exposeHeaders := c.convert(c.normalizeStrs(c.ExposeHeaders), http.CanonicalHeaderKey)
		headers.Set("Access-Control-Expose-Headers", strings.Join(exposeHeaders, ","))
	}
	if c.allowAllOrigins {
		headers.Set("Access-Control-Allow-Origin", "*")
	} else {
		headers.Set("Vary", "Origin")
	}
	return headers
}

func (c *cors) generatePreflightHeaders() http.Header {
	headers := make(http.Header)
	if c.AllowCredentials {
		headers.Set("Access-Control-Allow-Credentials", "true")
	}
	if len(c.AllowMethods) > 0 {
		allowMethods := c.convert(c.normalizeStrs(c.AllowMethods), strings.ToUpper)
		headers.Set("Access-Control-Allow-Methods", strings.Join(allowMethods, ","))
	}
	if len(c.AllowHeaders) > 0 {
		allowHeaders := c.convert(c.normalizeStrs(c.AllowHeaders), http.CanonicalHeaderKey)
		headers.Set("Access-Control-Allow-Headers", strings.Join(allowHeaders, ","))
	}
	if c.MaxAge > time.Duration(0) {
		headers.Set("Access-Control-Max-Age", fmt.Sprintf("%d", c.MaxAge/time.Second))
	}
	if c.allowAllOrigins {
		headers.Set("Access-Control-Allow-Origin", "*")
	} else {
		// If the server specifies an origin host rather than "*",
		// then it could also include Origin in the Vary response header
		// to indicate to clients that server responses will differ based
		// on the value of the Origin request header.
		headers.Add("Vary", "Origin")
		headers.Add("Vary", "Access-Control-Request-Method")
		headers.Add("Vary", "Access-Control-Request-Headers")
	}
	return headers
}

func (c *cors) convert(strs []string, f func(string) string) []string {
	var result []string
	for _, str := range strs {
		result = append(result, f(str))
	}
	return result
}

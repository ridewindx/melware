package melware

import (
	"net/http"
	"github.com/rs/cors"
	"github.com/ridewindx/mel"
)

type Cors struct {
	// AllowedOrigins is a list of origins a cross-domain request can be executed from.
	// If the special "*" value is present in the list, all origins will be allowed.
	// An origin may contain a wildcard (*) to replace 0 or more characters
	// (i.e.: http://*.domain.com). Usage of wildcards implies a small performance penalty.
	// Only one wildcard can be used per origin.
	// Default value is ["*"]
	AllowedOrigins []string
	// AllowOriginFunc is a custom function to validate the origin. It take the origin
	// as argument and returns true if allowed or false otherwise. If this option is
	// set, the content of AllowedOrigins is ignored.
	AllowOriginFunc func(origin string) bool
	// AllowedMethods is a list of methods the client is allowed to use with
	// cross-domain requests. Default value is simple methods (HEAD, GET and POST).
	AllowedMethods []string
	// AllowedHeaders is list of non simple headers the client is allowed to use with
	// cross-domain requests.
	// If the special "*" value is present in the list, all headers will be allowed.
	// Default value is [] but "Origin" is always appended to the list.
	AllowedHeaders []string
	// ExposedHeaders indicates which headers are safe to expose to the API of a CORS
	// API specification
	ExposedHeaders []string
	// AllowCredentials indicates whether the request can include user credentials like
	// cookies, HTTP authentication or client side SSL certificates.
	AllowCredentials bool
	// MaxAge indicates how long (in seconds) the results of a preflight request
	// can be cached
	MaxAge int
	// OptionsPassthrough instructs preflight to let other potential next handlers to
	// process the OPTIONS method. Turn this on if your application handles OPTIONS.
	OptionsPassthrough bool
	// Debugging flag adds additional output to debug server side CORS issues
	Debug bool

	*cors.Cors
}

func CorsAllowAll() *Cors {
	return &Cors{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"HEAD", "GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	}
}

func (c *Cors) Middleware() mel.Handler {
	if c.Cors == nil {
		c.Cors = cors.New(cors.Options{
			AllowedOrigins: c.AllowedOrigins,
			AllowOriginFunc: c.AllowOriginFunc,
			AllowedMethods: c.AllowedMethods,
			AllowedHeaders: c.AllowedHeaders,
			ExposedHeaders: c.ExposedHeaders,
			AllowCredentials: c.AllowCredentials,
			MaxAge: c.MaxAge,
			OptionsPassthrough: c.OptionsPassthrough,
			Debug: c.Debug,
		})
	}

	return func(ctx *mel.Context) {
		c.ServeHTTP(ctx.Writer, ctx.Request, func(writer http.ResponseWriter, request *http.Request) {
			ctx.Next()
		})
	}
}

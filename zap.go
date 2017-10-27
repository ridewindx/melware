package melware

import (
	"github.com/ridewindx/mel"
	"go.uber.org/zap"
	"time"
)

// Logrus returns a middleware that logs requests using logrus.
// Requests with errors are logged using logrus.Error().
// Requests without errors are logged using logrus.Info().
// The timeFormat specifies the time format string, e.g., time.RFC3339.
// The utc determines whether to use UTC time zone or local.
func Zap(logger *zap.SugaredLogger) mel.Handler {
	return func(c *mel.Context) {
		start := time.Now()

		// some evil middlewares modify this value
		path := c.Request.URL.Path
		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		fields := []interface{}{
			"status",     c.Writer.Status(),
			"method",     c.Request.Method,
			"path",       path,
			"ip",         c.ClientIP(),
			"latency",    latency,
			"user-agent", c.Request.UserAgent(),
		}

		if len(c.Errors) > 0 {
			// Append error string if this is an erroneous request.
			logger.Errorw(c.Errors.String(), fields...)
		} else {
			logger.Infow("", fields...)
		}
	}
}

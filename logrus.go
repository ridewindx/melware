package melware

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/ridewindx/mel"
)

// Logrus returns a middleware that logs requests using logrus.
// Requests with errors are logged using logrus.Error().
// Requests without errors are logged using logrus.Info().
// The timeFormat specifies the time format string, e.g., time.RFC3339.
// The utc determines whether to use UTC time zone or local.
func Logrus(logger *logrus.Logger, timeFormat string, utc bool) mel.Handler {
	return func(c *mel.Context) {
		start := time.Now()

		// some evil middlewares modify this value
		path := c.Request.URL.Path
		c.Next()

		end := time.Now()
		latency := end.Sub(start)
		if utc {
			end = end.UTC()
		}

		entry := logger.WithFields(logrus.Fields{
			"status":     c.Writer.Status(),
			"method":     c.Request.Method,
			"path":       path,
			"ip":         c.ClientIP(),
			"latency":    latency,
			"user-agent": c.Request.UserAgent(),
			"time":       end.Format(timeFormat),
		})

		if len(c.Errors) > 0 {
			// Append error string if this is an erroneous request.
			entry.Error(c.Errors.String())
		} else {
			entry.Info()
		}
	}
}

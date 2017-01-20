package melware

import (
	"compress/gzip"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/ridewindx/mel"
)

const (
	BestCompression    = gzip.BestCompression
	BestSpeed          = gzip.BestSpeed
	DefaultCompression = gzip.DefaultCompression
	NoCompression      = gzip.NoCompression
)

func Gzip(level int) mel.Handler {
	return func(c *mel.Context) {
		if !shouldCompress(c.Request) {
			return
		}
		gz, err := gzip.NewWriterLevel(c.Writer, level)
		if err != nil {
			panic(err)
		}

		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")
		c.Writer = &gzipWriter{c.Writer, gz}

		defer func() {
			c.Header("Content-Length", "0")
			gz.Close()
		}()

		c.Next()
	}
}

type gzipWriter struct {
	mel.ResponseWriter
	writer *gzip.Writer
}

func (g *gzipWriter) WriteString(s string) (int, error) {
	if !g.ResponseWriter.Written() {
		g.ResponseWriter.WriteHeader(g.ResponseWriter.Status())
	}
	return g.writer.Write([]byte(s))
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	if !g.ResponseWriter.Written() {
		g.ResponseWriter.WriteHeader(g.ResponseWriter.Status())
	}
	return g.writer.Write(data)
}

func shouldCompress(req *http.Request) bool {
	if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
		return false
	}
	ext := filepath.Ext(req.URL.Path)
	if len(ext) < 4 { // fast path
		return true
	}

	switch ext {
	case ".png", ".gif", ".jpeg", ".jpg":
		return false
	default:
		return true
	}
}

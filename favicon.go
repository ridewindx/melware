package melware

import (
	"io/ioutil"
	"strconv"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"github.com/ridewindx/mel"
)

const defaultMaxAge = 60 * 60 * 24 * 365; // 1 year

func Favicon(path string, maxAge ...int) mel.Handler {
	favicon, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	age := defaultMaxAge
	if len(maxAge) > 0 {
		age = maxAge[0]
	}

	hash := md5.New()
	io.WriteString(hash, string(favicon))
	etag := fmt.Sprintf("%x", hash.Sum(nil))

	return func(c *mel.Context) {
		req := c.Request
		if req.URL.Path != "/favicon.ico" {
			c.Next()
			return
		}

		if req.Method != "GET" && req.Method != "HEAD" {
			var status int
			if req.Method == "OPTIONS" {
				status = 200
			} else {
				status = 405
			}
			c.Header("Allow", "GET, HEAD, OPTIONS")
			c.Header("Content-Length", "0")
			c.Writer.WriteHeader(status)
			return
		}

		c.Header("Cache-Control", "public, max-age=" + strconv.Itoa(age))

		body := favicon
		if match, ok := req.Header["If-None-Match"]; ok {
			if match[0] == etag {
				c.Writer.WriteHeader(http.StatusNotModified)
				body = []byte{}
			}
		} else {
			c.Header("ETag", etag)
		}

		c.Writer.Write(body)
		return
	}
}

package cache

import (
    "time"
    "github.com/ridewindx/mel"
    "crypto/sha1"
    "io"
    "fmt"
    "net/http"
    "bytes"
    "log"
)

type responseCache struct {
    status int
    header http.Header
    body []byte
}

type cachedWriter struct {
    mel.ResponseWriter
    key string
    expire time.Duration
    *Cache
    error
    bytes.Buffer
}

func (w *cachedWriter) Write(bytes []byte) (int, error) {
    size, err := w.ResponseWriter.Write(bytes)
    if err == nil {
        w.Buffer.Write(bytes)
    } else {
        w.error = err
    }
    return size, err
}

func (w *cachedWriter) WriteString(s string) (int, error) {
    b := []byte(s)
    return w.Write(b)
}

func (w *cachedWriter) cache() {
    if w.error != nil {
        log.Printf("Do not cache key %s since: %s", w.key, w.error)
        return
    }
    rc := responseCache{
        status: w.Status(),
        header: w.Header(),
        body: w.Buffer.Bytes(),
    }
    err := w.Cache.Set(w.key, rc, w.expire)
    if err != nil {
        log.Printf("Cache key %s failed: %s", w.key, err)
    }
}

type Cache struct {
    KeyPrefix string
    Store
}

func (cache *Cache) CacheMiddleware(expire time.Duration) mel.Handler {
    return func(c *mel.Context) {

    }
}

func (cache *Cache) Cache(expire time.Duration, handler mel.Handler) mel.Handler {
    return func(c *mel.Context) {
        key := cache.makeKey(c.Request.URL.RequestURI())
        var r responseCache
        err := cache.Get(key, &r)
        if err != nil {
            cw := &cachedWriter{
                ResponseWriter: c.Writer,
                key: key,
                expire: expire,
                Cache: cache,
            }
            c.Writer = cw
            handler(c)
            cw.cache()
        } else {
            c.Status(r.status)
            for k, v := range r.header {
                c.Writer.Header()[k] = v
            }
            c.Writer.Write(r.body)
        }
    }
}

func (cache *Cache) makeKey(url string) string {
    key := url
    if len(key) > 200 {
        h := sha1.New()
        io.WriteString(h, key)
        key = string(h.Sum(nil))
    }
    return fmt.Sprintf("%s:%s", cache.KeyPrefix, key)
}

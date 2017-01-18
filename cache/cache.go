package cache

import (
    "time"
    "github.com/ridewindx/mel"
    "crypto/sha1"
    "io"
    "fmt"
    "net/http"
    "bytes"
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

    error

    bytes.Buffer
    *Cache
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

func (w *cachedWriter) cache() {
    if w.error == nil {
        // TODO
        return
    }
    rc := responseCache{
        status: w.Status(),
        header: w.Header(),
        body: w.Buffer.Bytes(),
    }
    err := w.Cache.Set(w.key, rc, w.expire)
    if err != nil {
        // TODO
    }
}

type Cache struct {
    Backend

    KeyPrefix string

    Store
}

func New(backend Backend) *Cache {
    return &Cache{
        Backend: backend,
    }
}

func (cache *Cache) CacheMiddleware(expire time.Duration) mel.Handler {
    return func(c *mel.Context) {

    }
}

func (cache *Cache) Cache(expire time.Duration, handler mel.Handler) mel.Handler {
    return func(c *mel.Context) {
        key := cache.makeKey(c.Request.URL.RequestURI())
        value, err := cache.Get(key)
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
            r := value.(responseCache)
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

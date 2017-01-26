package melware

import (
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/ridewindx/mel"
)

func newTestApp(cors *Cors) *mel.Mel {
    app := mel.New()
    app.Use(cors.Middleware())
    app.Get("/", func(c *mel.Context) {
        c.Text(200, "get")
    })
    app.Post("/", func(c *mel.Context) {
        c.Text(200, "post")
    })
    app.Patch("/", func(c *mel.Context) {
        c.Text(200, "patch")
    })
    return app
}

func performRequest(r http.Handler, method, origin string) *httptest.ResponseRecorder {
    req, _ := http.NewRequest(method, "/", nil)
    if len(origin) > 0 {
        req.Header.Set("Origin", origin)
    }
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)
    return w
}

func TestCorsDefaultConfig(t *testing.T) {
    cors := NewCors()
    cors.ExposeHeaders = []string{"exposed", "header", "hey"}

    assert.Equal(t, cors.AllowOrigins, []string{"*"})
    assert.Equal(t, cors.AllowMethods, []string{"GET", "POST", "HEAD"})
    assert.Equal(t, cors.AllowHeaders, []string{"Origin", "Accept", "Content-Type"})
    assert.Equal(t, cors.ExposeHeaders, []string{"exposed", "header", "hey"})
    assert.Equal(t, cors.AllowCredentials, false)
}

func TestCorsBadConfig(t *testing.T) {
    cors := NewCors()

    cors.AllowOrigins = []string{"*", "http://google.com"}
    assert.Panics(t, func() {
        cors.validateAllowOrigins()
    })

    cors.AllowOrigins = []string{"ftp://ftp.google.com"}
    assert.Panics(t, func() {
        cors.validateAllowOrigins()
    })

    cors.AllowOrigins = []string{"*"}
    cors.AllowOriginFunc = func(string) bool { return false }
    assert.Panics(t, func() {
        cors.validateAllowOrigins()
    })

    cors.AllowOrigins = nil
    cors.AllowOriginFunc = nil
    assert.Panics(t, func() {
        cors.validateAllowOrigins()
    })
}

func TestCorsNormalize(t *testing.T) {
    cors := NewCors()
    values := cors.normalizeStrs([]string{
        "http-Access ", "Post", "POST", " poSt  ",
        "HTTP-Access", "",
    })
    assert.Equal(t, values, []string{"http-access", "post", ""})

    values = cors.normalizeStrs(nil)
    assert.Nil(t, values)

    values = cors.normalizeStrs([]string{})
    assert.Nil(t, values)
}

func TestCorsConvert(t *testing.T) {
    cors := NewCors()
    methods := []string{"Get", "GET", "get"}
    headers := []string{"X-CSRF-TOKEN", "X-CSRF-Token", "x-csrf-token"}

    assert.Equal(t, []string{"GET", "GET", "GET"}, cors.convert(methods, strings.ToUpper))
    assert.Equal(t, []string{"X-Csrf-Token", "X-Csrf-Token", "X-Csrf-Token"}, cors.convert(headers, http.CanonicalHeaderKey))
}

func TestCorsGenerateNormalHeadersForAllowAllOrigins(t *testing.T) {
    cors := NewCors()
    cors.AllowOrigins = nil
    header := cors.generateNormalHeaders()
    assert.Equal(t, header.Get("Access-Control-Allow-Origin"), "")
    assert.Equal(t, header.Get("Vary"), "Origin")
    assert.Len(t, header, 1)

    cors.AllowOrigins = []string{"*"}
    cors.validateAllowOrigins()
    header = cors.generateNormalHeaders()
    assert.Equal(t, header.Get("Access-Control-Allow-Origin"), "*")
    assert.Equal(t, header.Get("Vary"), "")
    assert.Len(t, header, 1)
}

func TestCorsGenerateNormalHeadersForAllowCredentials(t *testing.T) {
    cors := NewCors()
    cors.AllowCredentials = true
    header := cors.generateNormalHeaders()
    assert.Equal(t, header.Get("Access-Control-Allow-Credentials"), "true")
    assert.Equal(t, header.Get("Vary"), "Origin")
    assert.Len(t, header, 2)
}

func TestCorsGenerateNormalHeadersForExposedHeaders(t *testing.T) {
    cors := NewCors()
    cors.ExposeHeaders = []string{"X-user", "xPassword"}
    header := cors.generateNormalHeaders()
    assert.Equal(t, header.Get("Access-Control-Expose-Headers"), "X-User,Xpassword")
    assert.Equal(t, header.Get("Vary"), "Origin")
    assert.Len(t, header, 2)
}

func TestCorsGeneratePreflightHeaders(t *testing.T) {
    cors := &Cors{}
    cors.AllowOrigins = nil
    header := cors.generatePreflightHeaders()
    assert.Equal(t, header.Get("Access-Control-Allow-Origin"), "")
    assert.Equal(t, header.Get("Vary"), "Origin")
    assert.Len(t, header, 1)

    cors.AllowOrigins = []string{"*"}
    cors.validateAllowOrigins()
    header = cors.generateNormalHeaders()
    assert.Equal(t, header.Get("Access-Control-Allow-Origin"), "*")
    assert.Equal(t, header.Get("Vary"), "")
    assert.Len(t, header, 1)
}

func TestCorsGeneratePreflightHeadersForAllowCredentials(t *testing.T) {
    cors := &Cors{}
    cors.AllowCredentials = true
    cors.AllowOrigins = nil
    header := cors.generatePreflightHeaders()
    assert.Equal(t, header.Get("Access-Control-Allow-Credentials"), "true")
    assert.Equal(t, header.Get("Vary"), "Origin")
    assert.Len(t, header, 2)
}

func TestCorsGeneratePreflightHeadersForAllowedMethods(t *testing.T) {
    cors := &Cors{}
    cors.AllowMethods = []string{"GET ", "post", "PUT", " put  "}
    header := cors.generatePreflightHeaders()
    assert.Equal(t, header.Get("Access-Control-Allow-Methods"), "GET,POST,PUT")
    assert.Equal(t, header.Get("Vary"), "Origin")
    assert.Len(t, header, 2)
}

func TestCorsGeneratePreflightHeadersForAllowedHeaders(t *testing.T) {
    cors := &Cors{}
    cors.AllowHeaders = []string{"X-user", "Content-Type"}
    header := cors.generatePreflightHeaders()
    assert.Equal(t, header.Get("Access-Control-Allow-Headers"), "X-User,Content-Type")
    assert.Equal(t, header.Get("Vary"), "Origin")
    assert.Len(t, header, 2)
}

func TestCorsGeneratePreflightHeadersForMaxAge(t *testing.T) {
    cors := &Cors{}
    cors.MaxAge = 12 * time.Hour
    header := cors.generatePreflightHeaders()
    assert.Equal(t, header.Get("Access-Control-Max-Age"), "43200") // 12*60*60
    assert.Equal(t, header.Get("Vary"), "Origin")
    assert.Len(t, header, 2)
}

func TestCorsValidateOrigin(t *testing.T) {
    cors := NewCors()
    cors.validateAllowOrigins()
    assert.True(t, cors.validateOrigin("http://google.com"))
    assert.True(t, cors.validateOrigin("https://google.com"))
    assert.True(t, cors.validateOrigin("example.com"))

    cors = NewCors()
    cors.AllowOrigins = []string{"https://google.com", "https://github.com"}
    cors.AllowOriginFunc = func(origin string) bool {
            return (origin == "http://news.ycombinator.com")
    }
    cors.validateAllowOrigins()

    assert.False(t, cors.validateOrigin("http://google.com"))
    assert.True(t, cors.validateOrigin("https://google.com"))
    assert.True(t, cors.validateOrigin("https://github.com"))
    assert.True(t, cors.validateOrigin("http://news.ycombinator.com"))
    assert.False(t, cors.validateOrigin("http://example.com"))
    assert.False(t, cors.validateOrigin("google.com"))
}

func TestCorsPassesAllowedOrigins(t *testing.T) {
    cors := NewCors()
    cors.AllowOrigins = []string{"http://google.com"}
    cors.AllowMethods =      []string{" GeT ", "get", "post", "PUT  ", "Head", "POST"}
    cors.AllowHeaders =     []string{"Content-type", "timeStamp "}
    cors.ExposeHeaders =   []string{"Data", "x-User"}
    cors.AllowCredentials = false
    cors.MaxAge =           12 * time.Hour
    cors.AllowOriginFunc = func(origin string) bool {
        return origin == "http://github.com"
    }

    router := newTestApp(cors)

    // no CORS request, origin == ""
    w := performRequest(router, "GET", "")
    assert.Equal(t, w.Body.String(), "get")
    assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
    assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))
    assert.Empty(t, w.Header().Get("Access-Control-Expose-Headers"))

    // allowed CORS request
    w = performRequest(router, "GET", "http://google.com")
    assert.Equal(t, w.Body.String(), "get")
    assert.Equal(t, w.Header().Get("Access-Control-Allow-Origin"), "http://google.com")
    assert.Equal(t, w.Header().Get("Access-Control-Allow-Credentials"), "")
    assert.Equal(t, w.Header().Get("Access-Control-Expose-Headers"), "Data,X-User")

    // deny CORS request
    w = performRequest(router, "GET", "https://google.com")
    assert.Equal(t, w.Code, 403)
    assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
    assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))
    assert.Empty(t, w.Header().Get("Access-Control-Expose-Headers"))

    // allowed CORS prefligh request
    w = performRequest(router, "OPTIONS", "http://github.com")
    assert.Equal(t, w.Code, 200)
    assert.Equal(t, w.Header().Get("Access-Control-Allow-Origin"), "http://github.com")
    assert.Equal(t, w.Header().Get("Access-Control-Allow-Credentials"), "")
    assert.Equal(t, w.Header().Get("Access-Control-Allow-Methods"), "GET,POST,PUT,HEAD")
    assert.Equal(t, w.Header().Get("Access-Control-Allow-Headers"), "Content-Type,Timestamp")
    assert.Equal(t, w.Header().Get("Access-Control-Max-Age"), "43200")

    // deny CORS prefligh request
    w = performRequest(router, "OPTIONS", "http://example.com")
    assert.Equal(t, w.Code, 403)
    assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
    assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))
    assert.Empty(t, w.Header().Get("Access-Control-Allow-Methods"))
    assert.Empty(t, w.Header().Get("Access-Control-Allow-Headers"))
    assert.Empty(t, w.Header().Get("Access-Control-Max-Age"))
}

func TestCorsPassesAllowedAllOrigins(t *testing.T) {
    cors := NewCors()
    cors.AllowOrigins = []string{"*"}
    cors.AllowMethods =      []string{" Patch ", "get", "post", "POST"}
    cors.AllowHeaders =     []string{"Content-type", "  testheader "}
    cors.ExposeHeaders =   []string{"Data2", "x-User2"}
    cors.AllowCredentials = false
    cors.MaxAge =           10 * time.Hour
    router := newTestApp(cors)

    // no CORS request, origin == ""
    w := performRequest(router, "GET", "")
    assert.Equal(t, w.Body.String(), "get")
    assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
    assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))
    assert.Empty(t, w.Header().Get("Access-Control-Expose-Headers"))

    // allowed CORS request
    w = performRequest(router, "POST", "example.com")
    assert.Equal(t, w.Body.String(), "post")
    assert.Equal(t, w.Header().Get("Access-Control-Allow-Origin"), "*")
    assert.Equal(t, w.Header().Get("Access-Control-Expose-Headers"), "Data2,X-User2")
    assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))

    // allowed CORS prefligh request
    w = performRequest(router, "OPTIONS", "https://facebook.com")
    assert.Equal(t, w.Code, 200)
    assert.Equal(t, w.Header().Get("Access-Control-Allow-Origin"), "*")
    assert.Equal(t, w.Header().Get("Access-Control-Allow-Methods"), "PATCH,GET,POST")
    assert.Equal(t, w.Header().Get("Access-Control-Allow-Headers"), "Content-Type,Testheader")
    assert.Equal(t, w.Header().Get("Access-Control-Max-Age"), "36000")
    assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))
}

package session

import (
	"log"
	"github.com/ridewindx/mel"
)

// Key for session storing into context.
// You can change it, i.e., session.ContextKey = "XXX".
var ContextKey  = "SESSION"

// Default key for flashes storing into session.
const flashesKey = "_flash"

// Options stores configuration for a session or session store.
// Fields are a subset of http.Cookie fields.
type Options struct {
	Path   string
	Domain string
	// MaxAge=0 means no 'Max-Age' attribute specified.
	// MaxAge<0 means delete cookie now, equivalently 'Max-Age: 0'.
	// MaxAge>0 means Max-Age attribute present and given in seconds.
	MaxAge   int
	Secure   bool
	HttpOnly bool
}

// Middleware returns a middleware that handles session.
func Middleware(name string, store Store) mel.Handler {
	return func(c *mel.Context) {
		s := &session{
			Name: name,
			Contents: make(Contents),
			store: store,
			context: c,
		}
		err := s.store.Get(c.Request, s.Name, s)
		if err != nil {
			log.Printf("session: %s\n", err)
		}
		c.Set(ContextKey, s)
		c.Next()
	}
}

// Session gets session for current request.
func Session(c *mel.Context) *session {
	return c.MustGet(ContextKey).(*session)
}

type Contents map[interface{}]interface{}

type session struct {
	// Session name.
	Name     string

	// Session ID.
	ID       string

	// Key-value pairs for holding your session contents.
	Contents

	// Whether changed after taken out.
	changed  bool

	store    Store

	context *mel.Context

	*Options
}

func (s *session) Get(key interface{}) (interface{}, bool) {
	return s.Contents[key]
}

func (s *session) Set(key interface{}, value interface{}) {
	s.Contents[key] = value
	s.changed = true
}

func (s *session) Delete(key interface{}) {
	delete(s.Contents, key)
	s.changed = true
}

func (s *session) Clear() {
	s.Contents = make(Contents)
	s.changed = true
}

func (s *session) AddFlash(args ...string) {
	key := flashesKey
	value := args[0]
	if len(args) > 1 {
		key = args[0]
		value = args[1]
	}
	var flashes []interface{}
	if v, ok := s.Contents[key]; ok {
		flashes = v.([]interface{})
	}
	s.Contents[key] = append(flashes, value)
}

func (s *session) Flashes(args ...string) []interface{} {
	key := flashesKey
	if len(args) > 0 {
		key = args[0]
	}
	if v, ok := s.Contents[key]; ok {
		delete(s.Contents, key)
		return v.([]interface{})
	}
	return nil
}

func (s *session) Save() error {
	if !s.changed {
		return nil
	}

	err := s.store.Save(s.context.Request, s.context.Writer, s)
	if err == nil {
		s.changed = false
	}
	return err
}

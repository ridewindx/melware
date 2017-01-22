package session

import (
	"github.com/ridewindx/mel"
)

// Default flashes key.
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

const (
	DefaultKey  = "github.com/gin-contrib/sessions"
)

func Sessions(name string, store Store) mel.Handler {
	return func(c *mel.Context) {
		s := &Session{
			Name: name,
			Values: make(SessionValues),
			store: store,
			Context: c,
		}
		c.Set(DefaultKey, s)
		c.Next()
	}
}

// shortcut to get session
func Default(c *mel.Context) Session {
	return c.MustGet(DefaultKey).(Session)
}

type SessionValues map[interface{}]interface{}

type Session struct {
	Name string
	ID string
	Values map[interface{}]interface{}
	changed bool
	store Store
	*mel.Context
	*Options
}

func (s *Session) Get(key interface{}) (interface{}, bool) {
	return s.Values[key]
}

func (s *Session) Set(key interface{}, value interface{}) {
	s.Values[key] = value
	s.changed = true
}

func (s *Session) Delete(key interface{}) {
	delete(s.Values, key)
	s.changed = true
}

func (s *Session) Clear() {
	s.Values = make(map[interface{}]interface{})
	s.changed = true
}

func (s *Session) AddFlash(args ...string) {
	key := flashesKey
	value := args[0]
	if len(args) > 1 {
		key = args[0]
		value = args[1]
	}
	var flashes []interface{}
	if v, ok := s.Values[key]; ok {
		flashes = v.([]interface{})
	}
	s.Values[key] = append(flashes, value)
}

func (s *Session) Flashes(args ...string) []interface{} {
	key := flashesKey
	if len(args) > 0 {
		key = args[0]
	}
	if v, ok := s.Values[key]; ok {
		delete(s.Values, key)
		return v.([]interface{})
	}
	return nil
}

func (s *Session) Save() error {
	if !s.changed {
		return nil
	}

	err := s.store.Save(s.Request, s.Writer, s)
	if err == nil {
		s.changed = false
	}
	return err
}

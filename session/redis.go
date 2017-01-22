package session

import (
	"encoding/base32"
	"errors"
	"net/http"
	"fmt"
	"bytes"
	"encoding/json"
	"encoding/gob"
	"strings"
	"time"
	"github.com/gorilla/securecookie"
	"github.com/garyburd/redigo/redis"
)

// Amount of time for cookies/redis keys to expire.
var sessionExpire = 86400 * 30

// SetMaxLength sets RedisStore.maxLength if the `l` argument is greater or equal 0
// maxLength restricts the maximum length of new sessions to l.
// If l is 0 there is no limit to the size of a session, use with caution.
// The default for a new RedisStore is 4096. Redis allows for max.
// value sizes of up to 512MB (http://redis.io/topics/data-types)
// Default: 4096,
func (s *RedisStore) SetMaxLength(l int) {
	if l >= 0 {
		s.maxLength = l
	}
}

// SetKeyPrefix set the prefix
func (s *RedisStore) SetKeyPrefix(p string) {
	s.keyPrefix = p
}

// SetSerializer sets the serializer
func (s *RedisStore) SetSerializer(ss SessionSerializer) {
	s.serializer = ss
}

// SetMaxAge restricts the maximum age, in seconds, of the session record
// both in database and a browser. This is to change session storage configuration.
// If you want just to remove session use your session `s` object and change it's
// `Options.MaxAge` to -1, as specified in
//    http://godoc.org/github.com/gorilla/sessions#Options
//
// Default is the one provided by this package value - `sessionExpire`.
// Set it to 0 for no restriction.
// Because we use `MaxAge` also in SecureCookie crypting algorithm you should
// use this function to change `MaxAge` value.
func (s *RedisStore) SetMaxAge(v int) {
	var c *securecookie.SecureCookie
	var ok bool
	s.Options.MaxAge = v
	for i := range s.Codecs {
		if c, ok = s.Codecs[i].(*securecookie.SecureCookie); ok {
			c.MaxAge(v)
		} else {
			fmt.Printf("Can't change MaxAge on codec %v\n", s.Codecs[i])
		}
	}
}

func dial(network, address, password string) (redis.Conn, error) {
	c, err := redis.Dial(network, address)
	if err != nil {
		return nil, err
	}
	if password != "" {
		if _, err := c.Do("AUTH", password); err != nil {
			c.Close()
			return nil, err
		}
	}
	return c, err
}

// NewRedisStore returns a new RedisStore.
// size: maximum number of idle connections.
func NewRedisStore(size int, network, address, password string, keyPairs ...[]byte) (*RedisStore, error) {
	return NewRedisStoreWithPool(&redis.Pool{
		MaxIdle:     size,
		IdleTimeout: 240 * time.Second,
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
		Dial: func() (redis.Conn, error) {
			return dial(network, address, password)
		},
	}, keyPairs...)
}

func dialWithDB(network, address, password, DB string) (redis.Conn, error) {
	c, err := dial(network, address, password)
	if err != nil {
		return nil, err
	}
	if _, err := c.Do("SELECT", DB); err != nil {
		c.Close()
		return nil, err
	}
	return c, err
}

// NewRedisStoreWithDB - like NewRedisStore but accepts `DB` parameter to select
// redis DB instead of using the default one ("0")
func NewRedisStoreWithDB(size int, network, address, password, DB string, keyPairs ...[]byte) (*RedisStore, error) {
	return NewRedisStoreWithPool(&redis.Pool{
		MaxIdle:     size,
		IdleTimeout: 240 * time.Second,
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
		Dial: func() (redis.Conn, error) {
			return dialWithDB(network, address, password, DB)
		},
	}, keyPairs...)
}

// NewRedisStoreWithPool instantiates a RedisStore with a *redis.Pool passed in.
func NewRedisStoreWithPool(pool *redis.Pool, keyPairs ...[]byte) (*RedisStore, error) {
	rs := &RedisStore{
		// http://godoc.org/github.com/garyburd/redigo/redis#Pool
		Pool:   pool,
		Codecs: securecookie.CodecsFromPairs(keyPairs...),
		Options: &Options{
			Path:   "/",
			MaxAge: sessionExpire,
		},
		DefaultMaxAge: 60 * 20, // 20 minutes seems like a reasonable default
		maxLength:     4096,
		keyPrefix:     "session_",
		serializer:    GobSerializer{},
	}
	_, err := rs.ping()
	return rs, err
}

// Close closes the underlying *redis.Pool
func (s *RedisStore) Close() error {
	return s.Pool.Close()
}

// ping does an internal ping against a server to check if it is alive.
func (s *RedisStore) ping() (bool, error) {
	conn := s.Pool.Get()
	defer conn.Close()
	data, err := conn.Do("PING")
	if err != nil || data == nil {
		return false, err
	}
	return (data == "PONG"), nil
}

// SessionSerializer provides an interface hook for alternative serializers
type SessionSerializer interface {
	Deserialize(d []byte, sv SessionValues) error
	Serialize(sv *SessionValues) ([]byte, error)
}

// JSONSerializer encode the session map to JSON.
type JSONSerializer struct{}

// Serialize to JSON. Will err if there are unmarshalable key values
func (s JSONSerializer) Serialize(sv SessionValues) ([]byte, error) {
	m := make(map[string]interface{}, len(sv))
	for k, v := range sv {
		ks, ok := k.(string)
		if !ok {
			err := fmt.Errorf("Non-string key value, cannot serialize session to JSON: %v", k)
			fmt.Printf("redistore.JSONSerializer.serialize() Error: %v", err)
			return nil, err
		}
		m[ks] = v
	}
	return json.Marshal(m)
}

// Deserialize back to map[string]interface{}
func (s JSONSerializer) Deserialize(d []byte, sv *SessionValues) error {
	m := make(map[string]interface{})
	err := json.Unmarshal(d, &m)
	if err != nil {
		fmt.Printf("redistore.JSONSerializer.deserialize() Error: %v", err)
		return err
	}
	for k, v := range m {
		sv[k] = v
	}
	return nil
}

// GobSerializer uses gob package to encode the session map
type GobSerializer struct{}

// Serialize using gob
func (s GobSerializer) Serialize(sv SessionValues) ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(sv)
	if err == nil {
		return buf.Bytes(), nil
	}
	return nil, err
}

// Deserialize back to map[interface{}]interface{}
func (s GobSerializer) Deserialize(d []byte, sv *SessionValues) error {
	dec := gob.NewDecoder(bytes.NewBuffer(d))
	return dec.Decode(sv)
}

type RedisStore struct {
	Pool          *redis.Pool
	Codecs        []securecookie.Codec
	Options       *Options // default configuration
	DefaultMaxAge int      // default Redis TTL for a MaxAge == 0 session
	maxLength     int
	keyPrefix     string
	serializer    SessionSerializer
}

func (s *RedisStore) Get(r *http.Request, name string, session *Session) error {
	cookie, err := r.Cookie(name)
	if err != nil {
		return err
	}

	err = securecookie.DecodeMulti(name, cookie.Value, &session.ID, s.Codecs...)
	if err != nil {
		return err
	}

	conn := s.Pool.Get()
	defer conn.Close()

	data, err := conn.Do("GET", s.keyPrefix+session.ID)
	if data == nil { // no data was associated with the key
		return nil
	}
	b, err := redis.Bytes(data, err)
	if err != nil {
		return err
	}
	err = s.serializer.Deserialize(b, &session.Values)
	return err
}

func (s *RedisStore) Save(r *http.Request, w http.ResponseWriter, session *Session) error {
	// Marked for deletion.
	if session.Options.MaxAge < 0 {
		conn := s.Pool.Get()
		defer conn.Close()
		if _, err := conn.Do("DEL", s.keyPrefix+session.ID); err != nil {
			return err
		}
		http.SetCookie(w, NewCookie(session.Name(), "", session.Options))
	} else {
		if session.ID == "" {
			id := securecookie.GenerateRandomKey(32)
			id = base32.StdEncoding.EncodeToString(id)
			session.ID = strings.TrimRight(id, "=")
		}

		b, err := s.serializer.Serialize(session.Values)
		if err != nil {
			return err
		}
		if s.maxLength != 0 && len(b) > s.maxLength {
			return errors.New("SessionStore: the value to store is too big")
		}
		conn := s.Pool.Get()
		defer conn.Close()

		age := session.Options.MaxAge
		if age == 0 {
			age = s.DefaultMaxAge
		}

		_, err = conn.Do("SETEX", s.keyPrefix+session.ID, age, b)
		if err != nil {
			return err
		}

		encoded, err := securecookie.EncodeMulti(session.Name(), session.ID, s.Codecs...)
		if err != nil {
			return err
		}
		http.SetCookie(w, NewCookie(session.Name(), encoded, session.Options))
	}
	return nil
}

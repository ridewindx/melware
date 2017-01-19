package cache

import (
	"time"
	"github.com/garyburd/redigo/redis"
	"bytes"
	"encoding/gob"
)

type RedisStore struct {
	pool *redis.Pool
	defaultExpiration time.Duration
}

var _ Store = &MemoryStore{}

func NewRedisStore(host, password string, defaultExpiration time.Duration) *RedisStore {
	var pool = &redis.Pool{
		MaxIdle:     5,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			// TODO: the redis protocol should probably be made settable
			c, err := redis.Dial("tcp", host)
			if err != nil {
				return nil, err
			}
			if len(password) > 0 {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			} else {
				if _, err := c.Do("PING"); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		// custom connection test method
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if _, err := c.Do("PING"); err != nil {
				return err
			}
			return nil
		},
	}
	return &RedisStore{pool, defaultExpiration}
}

func (c *RedisStore) Get(key string, ptr interface{}) error {
	conn := c.pool.Get()
	defer conn.Close()

	raw, err := conn.Do("GET", key)
	if raw == nil {
		return ErrCacheMiss
	}
	item, err := redis.Bytes(raw, err)
	if err != nil {
		return err
	}
	return deserialize(item, ptr)
}

func (c *RedisStore) Set(key string, value interface{}, expire time.Duration) error {
	b, err := serialize(value)
	if err != nil {
		return err
	}

	conn := c.pool.Get()
	defer conn.Close()

	if expire != FOREVER {
		if expire == DEFAULT {
			expire = c.defaultExpiration
		}
		_, err := conn.Do("SETEX", key, int32(expire/time.Second), b)
		return err
	} else {
		_, err = conn.Do("SET", key, b)
		return err
	}
}

func (c *RedisStore) Delete(key string) error {
	conn := c.pool.Get()
	defer conn.Close()

	exists, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	_, err = conn.Do("DEL", key)
	return err
}

func (c *RedisStore) Clear() error {
	conn := c.pool.Get()
	defer conn.Close()

	_, err := conn.Do("FLUSHALL")
	return err
}

// serialize returns a []byte representing the passed value
func serialize(value interface{}) ([]byte, error) {
	var b bytes.Buffer
	encoder := gob.NewEncoder(&b)
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// deserialize deserialices the passed []byte into a the passed ptr interface{}
func deserialize(data []byte, ptr interface{}) error {
	b := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(b)
	if err := decoder.Decode(ptr); err != nil {
		return err
	}
	return nil
}

package cache

import (
	"reflect"
	"time"
	memory "github.com/robfig/go-cache"
)

type MemoryStore struct {
	memory.Cache
}

var _ Store = &MemoryStore{}

func (c *MemoryStore) Get(key string, ptr interface{}) error {
	val, ok := c.Cache.Get(key)
	if !ok {
		return ErrCacheMiss
	}

	v := reflect.ValueOf(ptr)
	if !(v.Type().Kind() == reflect.Ptr && v.Elem().CanSet()) {
		panic("Underlying value of the interface is not a pointer")
	}

	v.Elem().Set(reflect.ValueOf(val))
	return nil
}

func (c *MemoryStore) Set(key string, value interface{}, expire time.Duration) error {
	c.Cache.Set(key, value, expire)
	return nil
}

func (c *MemoryStore) Delete(key string) error {
	c.Cache.Delete(key)
	return nil
}

func (c *MemoryStore) Clear() error {
	c.Cache.Flush()
	return nil
}

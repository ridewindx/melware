package cache

import (
    "time"
    "errors"
)

const (
    DEFAULT = time.Duration(0)
    FOREVER = time.Duration(-1)
)

var ErrCacheMiss = errors.New("cache missing")

type Store interface {
    // Get retrieves item from cache, and return nil.
    // If the key is not found, return ErrCacheMiss.
    // Value must be a pointer.
    Get(key string, ptr interface{}) error

    // Set sets item to cache.
    // If the key exists, replace the item.
    Set(key string, value interface{}, expire time.Duration) error

    // Delete removes item from cache.
    // If the key does not exist, do nothing.
    Delete(key string) error

    // Clear all items from
    Clear() error
}

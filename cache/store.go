package cache

import "time"

type Backend int

const (
    INMEMORY = iota
    REDIS
)

type Store interface {
    // Get retrieves item from cache, i.e., (item, true).
    // If the key is not found, return (nil, false).
    Get(key string) (interface{}, error)

    // Set sets item to cache.
    // If the key exists, replace the item.
    Set(key string, value interface{}, expire time.Duration) error

    // Delete removes item from cache.
    // If the key does not exist, do nothing.
    Delete(key string) error

    // Clear all items from cache.
    Clear() error
}

package cache

import (
	"context"
	"time"
)

// Cache defines the interface for cache operations
type Cache interface {
	// Get retrieves a value from the cache
	Get(ctx context.Context, key string, dest interface{}) error

	// Set stores a value in the cache with TTL
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Delete removes a key from the cache
	Delete(ctx context.Context, keys ...string) error

	// DeletePattern removes keys matching a pattern
	DeletePattern(ctx context.Context, pattern string) error

	// Exists checks if a key exists
	Exists(ctx context.Context, key string) (bool, error)

	// Close closes the cache connection
	Close() error
}

// ErrCacheMiss is returned when a key is not found in the cache
type ErrCacheMiss struct {
	Key string
}

func (e *ErrCacheMiss) Error() string {
	return "cache miss: " + e.Key
}

// IsCacheMiss checks if an error is a cache miss
func IsCacheMiss(err error) bool {
	_, ok := err.(*ErrCacheMiss)
	return ok
}

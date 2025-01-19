// Package stream provides functionality for processing media streams and caching.
package stream

import (
	"github.com/allegro/bigcache"
	"time"
)

// Cache wraps the bigcache instance for easier use.
type Cache struct {
	cache *bigcache.BigCache
}

// NewCache initializes a new Cache instance with the specified expiration time.
func NewCache(expiration time.Duration) (*Cache, error) {
	config := bigcache.Config{
		Shards:             4096,           // Number of cache shards (divisions for concurrency)
		LifeWindow:         expiration,     // Cache expiration duration
		CleanWindow:        expiration / 2, // Frequency of clean-up tasks
		MaxEntriesInWindow: 1000 * 10 * 60, // Max entries in the given time window
		MaxEntrySize:       500,            // Max size of a single entry
		Verbose:            false,          // Toggle additional logging
		HardMaxCacheSize:   0,              // Set max cache size in MB (0 means unlimited)
	}

	cacheInstance, err := bigcache.NewBigCache(config)
	if err != nil {
		return nil, err
	}

	return &Cache{cache: cacheInstance}, nil
}

// Set adds a new key-value pair to the cache.
func (c *Cache) Set(key string, value string) error {
	return c.cache.Set(key, []byte(value))
}

// Get retrieves the value associated with the key from the cache.
// Returns the value and a boolean indicating whether the key was found.
func (c *Cache) Get(key string) (string, bool) {
	value, err := c.cache.Get(key)
	if err != nil {
		return "", false
	}
	return string(value), true
}

// Delete removes a key-value pair from the cache.
func (c *Cache) Delete(key string) error {
	return c.cache.Delete(key)
}

// Cleanup clears all expired entries from the cache.
func (c *Cache) Cleanup() {
	// bigcache handles cleanup automatically based on the configured LifeWindow.
}

package cache

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	// ErrKeyNotFound is returned when a key is not found in cache
	ErrKeyNotFound = errors.New("key not found")
)

// cacheEntry represents a single cache entry with expiration
type cacheEntry struct {
	value     []byte
	expiresAt time.Time
}

// isExpired checks if the entry has expired
func (e *cacheEntry) isExpired() bool {
	if e.expiresAt.IsZero() {
		return false // No expiration
	}
	return time.Now().After(e.expiresAt)
}

// memoryCache implements Cache using an in-memory map
type memoryCache struct {
	mu      sync.RWMutex
	entries map[string]*cacheEntry
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache() Cache {
	return &memoryCache{
		entries: make(map[string]*cacheEntry),
	}
}

// Get retrieves a value by key
func (c *memoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.entries[key]
	if !ok {
		return nil, ErrKeyNotFound
	}

	if entry.isExpired() {
		return nil, ErrKeyNotFound
	}

	// Return a copy to prevent modification
	value := make([]byte, len(entry.value))
	copy(value, entry.value)
	return value, nil
}

// Set stores a value with optional TTL (0 = no expiration)
func (c *memoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Make a copy of the value
	valueCopy := make([]byte, len(value))
	copy(valueCopy, value)

	entry := &cacheEntry{
		value: valueCopy,
	}

	if ttl > 0 {
		entry.expiresAt = time.Now().Add(ttl)
	}

	c.entries[key] = entry
	return nil
}

// Delete removes a key
func (c *memoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, key)
	return nil
}

// Exists checks if key exists
func (c *memoryCache) Exists(ctx context.Context, key string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.entries[key]
	if !ok {
		return false, nil
	}

	if entry.isExpired() {
		return false, nil
	}

	return true, nil
}

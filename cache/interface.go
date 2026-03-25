package cache

import (
	"context"
	"time"
)

// Cache abstracts short-lived data storage (context tokens, configs, etc.)
type Cache interface {
	// Get retrieves a value by key
	Get(ctx context.Context, key string) ([]byte, error)

	// Set stores a value with optional TTL (0 = no expiration)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Delete removes a key
	Delete(ctx context.Context, key string) error

	// Exists checks if key exists
	Exists(ctx context.Context, key string) (bool, error)
}

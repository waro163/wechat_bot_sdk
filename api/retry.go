package api

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// RetryConfig controls retry behavior
type RetryConfig struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
}

// DefaultRetryConfig returns sensible defaults
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 2 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
	}
}

// Retry executes fn with exponential backoff
func Retry(ctx context.Context, cfg *RetryConfig, fn func() error) error {
	if cfg == nil {
		cfg = DefaultRetryConfig()
	}

	var lastErr error
	delay := cfg.InitialDelay

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		lastErr = fn()

		if lastErr == nil {
			return nil
		}

		// Don't retry if not retryable
		if !isRetryable(lastErr) {
			return lastErr
		}

		// Don't sleep after last attempt
		if attempt < cfg.MaxAttempts {
			// Add jitter to prevent thundering herd
			jitter := time.Duration(rand.Int63n(int64(delay / 10)))
			sleepTime := delay + jitter

			select {
			case <-time.After(sleepTime):
				// Calculate next delay
				delay = time.Duration(float64(delay) * cfg.Multiplier)
				if delay > cfg.MaxDelay {
					delay = cfg.MaxDelay
				}
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return fmt.Errorf("max retries (%d) exceeded: %w", cfg.MaxAttempts, lastErr)
}

// isRetryable determines if an error should trigger a retry
// This is a simple implementation - can be enhanced based on error types
func isRetryable(err error) bool {
	// For now, consider all errors retryable
	// In production, you'd check for specific error types
	// (network errors, 5xx errors, etc.)
	return err != nil
}

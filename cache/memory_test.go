package cache

import (
	"context"
	"testing"
	"time"
)

// TestMemoryCache_SetAndGet tests basic set and get operations
func TestMemoryCache_SetAndGet(t *testing.T) {
	cache := NewMemoryCache()
	ctx := context.Background()

	key := "test-key"
	value := []byte("test value")

	// Set value
	if err := cache.Set(ctx, key, value, 1*time.Hour); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Get value
	got, err := cache.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if string(got) != string(value) {
		t.Errorf("Value mismatch: got %q, want %q", got, value)
	}
}

// TestMemoryCache_GetNonExistent tests getting non-existent key
func TestMemoryCache_GetNonExistent(t *testing.T) {
	cache := NewMemoryCache()
	ctx := context.Background()

	_, err := cache.Get(ctx, "non-existent")
	if err != ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound, got %v", err)
	}
}

// TestMemoryCache_Update tests updating existing key
func TestMemoryCache_Update(t *testing.T) {
	cache := NewMemoryCache()
	ctx := context.Background()

	key := "update-key"

	// Set initial value
	if err := cache.Set(ctx, key, []byte("value1"), 1*time.Hour); err != nil {
		t.Fatalf("Initial Set failed: %v", err)
	}

	// Update value
	if err := cache.Set(ctx, key, []byte("value2"), 1*time.Hour); err != nil {
		t.Fatalf("Update Set failed: %v", err)
	}

	// Get and verify updated value
	got, err := cache.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if string(got) != "value2" {
		t.Errorf("Value not updated: got %q, want %q", got, "value2")
	}
}

// TestMemoryCache_Delete tests deletion
func TestMemoryCache_Delete(t *testing.T) {
	cache := NewMemoryCache()
	ctx := context.Background()

	key := "delete-key"
	value := []byte("to be deleted")

	// Set value
	if err := cache.Set(ctx, key, value, 1*time.Hour); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Delete
	if err := cache.Delete(ctx, key); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify it's gone
	_, err := cache.Get(ctx, key)
	if err != ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound after delete, got %v", err)
	}
}

// TestMemoryCache_DeleteNonExistent tests deleting non-existent key
func TestMemoryCache_DeleteNonExistent(t *testing.T) {
	cache := NewMemoryCache()
	ctx := context.Background()

	// Should not error
	if err := cache.Delete(ctx, "non-existent"); err != nil {
		t.Errorf("Delete non-existent key should not error: %v", err)
	}
}

// TestMemoryCache_Exists tests key existence check
func TestMemoryCache_Exists(t *testing.T) {
	cache := NewMemoryCache()
	ctx := context.Background()

	key := "exists-key"

	// Check non-existent key
	exists, err := cache.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Error("Key should not exist yet")
	}

	// Set value
	if err := cache.Set(ctx, key, []byte("value"), 1*time.Hour); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Check existing key
	exists, err = cache.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("Key should exist")
	}

	// Delete and check again
	cache.Delete(ctx, key)
	exists, err = cache.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Error("Key should not exist after delete")
	}
}

// TestMemoryCache_TTLExpiration tests that entries expire after TTL
func TestMemoryCache_TTLExpiration(t *testing.T) {
	cache := NewMemoryCache()
	ctx := context.Background()

	key := "ttl-key"
	value := []byte("expires soon")

	// Set with short TTL
	if err := cache.Set(ctx, key, value, 100*time.Millisecond); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Should exist immediately
	got, err := cache.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if string(got) != string(value) {
		t.Errorf("Value mismatch: got %q, want %q", got, value)
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired now
	_, err = cache.Get(ctx, key)
	if err != ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound after TTL expiration, got %v", err)
	}
}

// TestMemoryCache_ZeroTTL tests setting with zero TTL (no expiration)
func TestMemoryCache_ZeroTTL(t *testing.T) {
	cache := NewMemoryCache()
	ctx := context.Background()

	key := "zero-ttl-key"
	value := []byte("never expires")

	// Set with zero TTL
	if err := cache.Set(ctx, key, value, 0); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Should still exist after some time
	time.Sleep(100 * time.Millisecond)

	got, err := cache.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if string(got) != string(value) {
		t.Errorf("Value mismatch: got %q, want %q", got, value)
	}
}

// TestMemoryCache_MultipleKeys tests handling multiple keys
func TestMemoryCache_MultipleKeys(t *testing.T) {
	cache := NewMemoryCache()
	ctx := context.Background()

	keys := []string{"key1", "key2", "key3", "key4", "key5"}
	values := [][]byte{[]byte("value1"), []byte("value2"), []byte("value3"), []byte("value4"), []byte("value5")}

	// Set all keys
	for i, key := range keys {
		if err := cache.Set(ctx, key, values[i], 1*time.Hour); err != nil {
			t.Fatalf("Set(%s) failed: %v", key, err)
		}
	}

	// Get all keys and verify
	for i, key := range keys {
		got, err := cache.Get(ctx, key)
		if err != nil {
			t.Fatalf("Get(%s) failed: %v", key, err)
		}
		if string(got) != string(values[i]) {
			t.Errorf("Value mismatch for %s: got %q, want %q", key, got, values[i])
		}
	}

	// Delete one key
	if err := cache.Delete(ctx, "key3"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify other keys still exist
	for _, key := range []string{"key1", "key2", "key4", "key5"} {
		exists, err := cache.Exists(ctx, key)
		if err != nil {
			t.Fatalf("Exists(%s) failed: %v", key, err)
		}
		if !exists {
			t.Errorf("Key %s should still exist", key)
		}
	}

	// Verify deleted key is gone
	exists, err := cache.Exists(ctx, "key3")
	if err != nil {
		t.Fatalf("Exists(key3) failed: %v", err)
	}
	if exists {
		t.Error("Key key3 should not exist")
	}
}

// TestMemoryCache_EmptyValue tests storing empty byte slices
func TestMemoryCache_EmptyValue(t *testing.T) {
	cache := NewMemoryCache()
	ctx := context.Background()

	key := "empty-key"
	value := []byte{}

	// Set empty value
	if err := cache.Set(ctx, key, value, 1*time.Hour); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Get and verify
	got, err := cache.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if len(got) != 0 {
		t.Errorf("Expected empty value, got %d bytes", len(got))
	}
}

// TestMemoryCache_LargeValue tests storing large values
func TestMemoryCache_LargeValue(t *testing.T) {
	cache := NewMemoryCache()
	ctx := context.Background()

	key := "large-key"
	value := make([]byte, 1024*1024) // 1MB
	for i := range value {
		value[i] = byte(i % 256)
	}

	// Set large value
	if err := cache.Set(ctx, key, value, 1*time.Hour); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Get and verify
	got, err := cache.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if len(got) != len(value) {
		t.Errorf("Length mismatch: got %d, want %d", len(got), len(value))
	}

	// Verify content matches
	for i := range got {
		if got[i] != value[i] {
			t.Errorf("Byte mismatch at index %d: got %d, want %d", i, got[i], value[i])
			break
		}
	}
}

// TestMemoryCache_ConcurrentAccess tests concurrent operations
func TestMemoryCache_ConcurrentAccess(t *testing.T) {
	cache := NewMemoryCache()
	ctx := context.Background()

	done := make(chan bool, 20)

	// Concurrent writes
	for i := range 10 {
		go func(n int) {
			key := "concurrent-key"
			value := []byte{byte(n)}
			cache.Set(ctx, key, value, 1*time.Hour)
			done <- true
		}(i)
	}

	// Concurrent reads
	for range 10 {
		go func() {
			cache.Get(ctx, "concurrent-key")
			done <- true
		}()
	}

	// Wait for all goroutines
	for range 20 {
		<-done
	}

	// Should not panic and should have a value
	_, err := cache.Get(ctx, "concurrent-key")
	if err != nil && err != ErrKeyNotFound {
		t.Errorf("Unexpected error after concurrent access: %v", err)
	}
}

// TestMemoryCache_UpdateTTL tests updating TTL on existing entry
func TestMemoryCache_UpdateTTL(t *testing.T) {
	cache := NewMemoryCache()
	ctx := context.Background()

	key := "update-ttl-key"
	value := []byte("value")

	// Set with short TTL
	if err := cache.Set(ctx, key, value, 100*time.Millisecond); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Wait a bit but not long enough to expire
	time.Sleep(50 * time.Millisecond)

	// Update with longer TTL
	if err := cache.Set(ctx, key, value, 1*time.Hour); err != nil {
		t.Fatalf("Update Set failed: %v", err)
	}

	// Wait past original TTL
	time.Sleep(100 * time.Millisecond)

	// Should still exist because TTL was extended
	got, err := cache.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if string(got) != string(value) {
		t.Errorf("Value mismatch: got %q, want %q", got, value)
	}
}

// TestMemoryCache_ContextCancellation tests behavior with cancelled context
func TestMemoryCache_ContextCancellation(t *testing.T) {
	cache := NewMemoryCache()

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Operations should still work (context is passed but not used in memory cache)
	key := "ctx-key"
	value := []byte("value")

	if err := cache.Set(ctx, key, value, 1*time.Hour); err != nil {
		t.Fatalf("Set with cancelled context failed: %v", err)
	}

	if _, err := cache.Get(ctx, key); err != nil {
		t.Fatalf("Get with cancelled context failed: %v", err)
	}
}

// BenchmarkMemoryCache_Set benchmarks set operations
func BenchmarkMemoryCache_Set(b *testing.B) {
	cache := NewMemoryCache()
	ctx := context.Background()
	value := []byte("benchmark value")

	b.ResetTimer()
	for range b.N {
		cache.Set(ctx, "bench-key", value, 1*time.Hour)
	}
}

// BenchmarkMemoryCache_Get benchmarks get operations
func BenchmarkMemoryCache_Get(b *testing.B) {
	cache := NewMemoryCache()
	ctx := context.Background()
	value := []byte("benchmark value")
	cache.Set(ctx, "bench-key", value, 1*time.Hour)

	b.ResetTimer()
	for range b.N {
		cache.Get(ctx, "bench-key")
	}
}

// BenchmarkMemoryCache_SetGet benchmarks combined set and get
func BenchmarkMemoryCache_SetGet(b *testing.B) {
	cache := NewMemoryCache()
	ctx := context.Background()
	value := []byte("benchmark value")

	b.ResetTimer()
	for range b.N {
		cache.Set(ctx, "bench-key", value, 1*time.Hour)
		cache.Get(ctx, "bench-key")
	}
}

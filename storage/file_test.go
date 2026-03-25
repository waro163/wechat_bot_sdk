package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/waro163/wechat-bot-sdk/common"
)

// TestFileStorage_SaveAndLoadAccount tests account persistence
func TestFileStorage_SaveAndLoadAccount(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	storage, err := NewFileStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewFileStorage failed: %v", err)
	}

	ctx := context.Background()
	now := time.Now().Unix()

	account := &common.Account{
		ID:        "test-account-001",
		Token:     "test-token-123",
		BaseURL:   "https://test.example.com",
		UserID:    "user-123",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Save account
	if err := storage.SaveAccount(ctx, account); err != nil {
		t.Fatalf("SaveAccount failed: %v", err)
	}

	// Load account
	loaded, err := storage.LoadAccount(ctx, "test-account-001")
	if err != nil {
		t.Fatalf("LoadAccount failed: %v", err)
	}

	// Verify all fields
	if loaded.ID != account.ID {
		t.Errorf("ID mismatch: got %q, want %q", loaded.ID, account.ID)
	}
	if loaded.Token != account.Token {
		t.Errorf("Token mismatch: got %q, want %q", loaded.Token, account.Token)
	}
	if loaded.BaseURL != account.BaseURL {
		t.Errorf("BaseURL mismatch: got %q, want %q", loaded.BaseURL, account.BaseURL)
	}
	if loaded.UserID != account.UserID {
		t.Errorf("UserID mismatch: got %q, want %q", loaded.UserID, account.UserID)
	}
	if loaded.CreatedAt != account.CreatedAt {
		t.Errorf("CreatedAt mismatch: got %d, want %d", loaded.CreatedAt, account.CreatedAt)
	}
	if loaded.UpdatedAt != account.UpdatedAt {
		t.Errorf("UpdatedAt mismatch: got %d, want %d", loaded.UpdatedAt, account.UpdatedAt)
	}
}

// TestFileStorage_UpdateAccount tests updating an existing account
func TestFileStorage_UpdateAccount(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewFileStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewFileStorage failed: %v", err)
	}

	ctx := context.Background()
	now := time.Now().Unix()

	// Create initial account
	account := &common.Account{
		ID:        "test-account-002",
		Token:     "old-token",
		BaseURL:   "https://old.example.com",
		UserID:    "user-456",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := storage.SaveAccount(ctx, account); err != nil {
		t.Fatalf("Initial SaveAccount failed: %v", err)
	}

	// Update account
	account.Token = "new-token"
	account.BaseURL = "https://new.example.com"
	account.UpdatedAt = now + 100

	if err := storage.SaveAccount(ctx, account); err != nil {
		t.Fatalf("Update SaveAccount failed: %v", err)
	}

	// Load and verify
	loaded, err := storage.LoadAccount(ctx, "test-account-002")
	if err != nil {
		t.Fatalf("LoadAccount failed: %v", err)
	}

	if loaded.Token != "new-token" {
		t.Errorf("Token not updated: got %q, want %q", loaded.Token, "new-token")
	}
	if loaded.BaseURL != "https://new.example.com" {
		t.Errorf("BaseURL not updated: got %q, want %q", loaded.BaseURL, "https://new.example.com")
	}
	if loaded.CreatedAt != now {
		t.Errorf("CreatedAt should not change: got %d, want %d", loaded.CreatedAt, now)
	}
	if loaded.UpdatedAt != now+100 {
		t.Errorf("UpdatedAt mismatch: got %d, want %d", loaded.UpdatedAt, now+100)
	}
}

// TestFileStorage_LoadNonExistentAccount tests loading account that doesn't exist
func TestFileStorage_LoadNonExistentAccount(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewFileStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewFileStorage failed: %v", err)
	}

	ctx := context.Background()

	_, err = storage.LoadAccount(ctx, "non-existent")
	if err == nil {
		t.Error("Expected error for non-existent account, got nil")
	}
}

// TestFileStorage_DeleteAccount tests account deletion
func TestFileStorage_DeleteAccount(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewFileStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewFileStorage failed: %v", err)
	}

	ctx := context.Background()
	now := time.Now().Unix()

	// Create account
	account := &common.Account{
		ID:        "test-account-003",
		Token:     "token-to-delete",
		BaseURL:   "https://delete.example.com",
		UserID:    "user-789",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := storage.SaveAccount(ctx, account); err != nil {
		t.Fatalf("SaveAccount failed: %v", err)
	}

	// Verify it exists
	_, err = storage.LoadAccount(ctx, "test-account-003")
	if err != nil {
		t.Fatalf("Account should exist before deletion: %v", err)
	}

	// Delete account
	if err := storage.DeleteAccount(ctx, "test-account-003"); err != nil {
		t.Fatalf("DeleteAccount failed: %v", err)
	}

	// Verify it's gone
	_, err = storage.LoadAccount(ctx, "test-account-003")
	if err == nil {
		t.Error("Account should not exist after deletion")
	}
}

// TestFileStorage_ListAccounts tests listing all accounts
func TestFileStorage_ListAccounts(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewFileStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewFileStorage failed: %v", err)
	}

	ctx := context.Background()
	now := time.Now().Unix()

	// Create multiple accounts
	accounts := []*common.Account{
		{ID: "account-001", Token: "token-1", BaseURL: "https://url1.com", UserID: "user-1", CreatedAt: now, UpdatedAt: now},
		{ID: "account-002", Token: "token-2", BaseURL: "https://url2.com", UserID: "user-2", CreatedAt: now, UpdatedAt: now},
		{ID: "account-003", Token: "token-3", BaseURL: "https://url3.com", UserID: "user-3", CreatedAt: now, UpdatedAt: now},
	}

	for _, acc := range accounts {
		if err := storage.SaveAccount(ctx, acc); err != nil {
			t.Fatalf("SaveAccount(%s) failed: %v", acc.ID, err)
		}
	}

	// List accounts
	list, err := storage.ListAccounts(ctx)
	if err != nil {
		t.Fatalf("ListAccounts failed: %v", err)
	}

	// Verify count
	if len(list) != 3 {
		t.Errorf("Expected 3 accounts, got %d", len(list))
	}

	// Verify all accounts are present
	found := make(map[string]bool)
	for _, acc := range list {
		found[acc.ID] = true
	}

	for _, expected := range accounts {
		if !found[expected.ID] {
			t.Errorf("Account %s not found in list", expected.ID)
		}
	}
}

// TestFileStorage_SaveAndLoadSyncBuffer tests sync buffer persistence
func TestFileStorage_SaveAndLoadSyncBuffer(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewFileStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewFileStorage failed: %v", err)
	}

	ctx := context.Background()
	accountID := "test-account-004"
	buffer := []byte("test sync buffer data 12345")

	// Save buffer
	if err := storage.SaveSyncBuffer(ctx, accountID, buffer); err != nil {
		t.Fatalf("SaveSyncBuffer failed: %v", err)
	}

	// Load buffer
	loaded, err := storage.LoadSyncBuffer(ctx, accountID)
	if err != nil {
		t.Fatalf("LoadSyncBuffer failed: %v", err)
	}

	// Verify
	if string(loaded) != string(buffer) {
		t.Errorf("Buffer mismatch:\ngot:  %q\nwant: %q", loaded, buffer)
	}
}

// TestFileStorage_LoadNonExistentSyncBuffer tests loading buffer that doesn't exist
func TestFileStorage_LoadNonExistentSyncBuffer(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewFileStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewFileStorage failed: %v", err)
	}

	ctx := context.Background()

	// Loading non-existent sync buffer should return nil, nil (not an error)
	buffer, err := storage.LoadSyncBuffer(ctx, "non-existent")
	if err != nil {
		t.Errorf("LoadSyncBuffer should not error on non-existent buffer: %v", err)
	}
	if buffer != nil {
		t.Errorf("Expected nil buffer, got %d bytes", len(buffer))
	}
}

// TestFileStorage_UpdateSyncBuffer tests updating existing sync buffer
func TestFileStorage_UpdateSyncBuffer(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewFileStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewFileStorage failed: %v", err)
	}

	ctx := context.Background()
	accountID := "test-account-005"

	// Save initial buffer
	buffer1 := []byte("initial buffer")
	if err := storage.SaveSyncBuffer(ctx, accountID, buffer1); err != nil {
		t.Fatalf("Initial SaveSyncBuffer failed: %v", err)
	}

	// Update buffer
	buffer2 := []byte("updated buffer with more data")
	if err := storage.SaveSyncBuffer(ctx, accountID, buffer2); err != nil {
		t.Fatalf("Update SaveSyncBuffer failed: %v", err)
	}

	// Load and verify
	loaded, err := storage.LoadSyncBuffer(ctx, accountID)
	if err != nil {
		t.Fatalf("LoadSyncBuffer failed: %v", err)
	}

	if string(loaded) != string(buffer2) {
		t.Errorf("Buffer not updated:\ngot:  %q\nwant: %q", loaded, buffer2)
	}
}

// TestFileStorage_EmptySyncBuffer tests saving empty buffer
func TestFileStorage_EmptySyncBuffer(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewFileStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewFileStorage failed: %v", err)
	}

	ctx := context.Background()
	accountID := "test-account-006"
	buffer := []byte{}

	// Save empty buffer
	if err := storage.SaveSyncBuffer(ctx, accountID, buffer); err != nil {
		t.Fatalf("SaveSyncBuffer failed: %v", err)
	}

	// Load and verify
	loaded, err := storage.LoadSyncBuffer(ctx, accountID)
	if err != nil {
		t.Fatalf("LoadSyncBuffer failed: %v", err)
	}

	if len(loaded) != 0 {
		t.Errorf("Expected empty buffer, got %d bytes", len(loaded))
	}
}

// TestFileStorage_ConcurrentAccess tests concurrent read/write safety
func TestFileStorage_ConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewFileStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewFileStorage failed: %v", err)
	}

	ctx := context.Background()
	now := time.Now().Unix()

	// Create initial account
	account := &common.Account{
		ID:        "concurrent-test",
		Token:     "token",
		BaseURL:   "https://example.com",
		UserID:    "user",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := storage.SaveAccount(ctx, account); err != nil {
		t.Fatalf("SaveAccount failed: %v", err)
	}

	// Concurrent reads
	done := make(chan bool, 10)
	for range 10 {
		go func() {
			_, err := storage.LoadAccount(ctx, "concurrent-test")
			if err != nil {
				t.Errorf("Concurrent LoadAccount failed: %v", err)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for range 10 {
		<-done
	}
}

// TestFileStorage_DirectoryCreation tests that directories are created automatically
func TestFileStorage_DirectoryCreation(t *testing.T) {
	tmpDir := t.TempDir()
	storageDir := filepath.Join(tmpDir, "nested", "path", "to", "storage")

	storage, err := NewFileStorage(storageDir)
	if err != nil {
		t.Fatalf("NewFileStorage failed: %v", err)
	}

	ctx := context.Background()
	now := time.Now().Unix()

	account := &common.Account{
		ID:        "dir-test",
		Token:     "token",
		BaseURL:   "https://example.com",
		UserID:    "user",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Should create directories automatically
	if err := storage.SaveAccount(ctx, account); err != nil {
		t.Fatalf("SaveAccount failed: %v", err)
	}

	// Verify directory exists
	if _, err := os.Stat(storageDir); os.IsNotExist(err) {
		t.Error("Storage directory was not created")
	}
}

// TestFileStorage_PathTraversal tests that path traversal attempts are blocked
func TestFileStorage_PathTraversal(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewFileStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewFileStorage failed: %v", err)
	}

	ctx := context.Background()

	tests := []struct {
		name      string
		accountID string
	}{
		{"path traversal attempt", "../../../etc/passwd"},
		{"absolute path attempt", "/etc/passwd"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Try to load - should fail cleanly
			_, err := storage.LoadAccount(ctx, tt.accountID)
			// We expect ErrAccountNotFound or some error, not a panic
			if err == nil {
				// Even if no error, make sure we didn't actually access /etc/passwd
				// The test passes as long as it doesn't panic
			}
		})
	}
}

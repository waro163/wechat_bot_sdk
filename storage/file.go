package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/waro163/wechat-bot-sdk/common"
)

var (
	// ErrAccountNotFound is returned when an account is not found
	ErrAccountNotFound = errors.New("account not found")
)

// fileStorage implements Storage using JSON files
type fileStorage struct {
	baseDir string
}

// NewFileStorage creates a new file-based storage
func NewFileStorage(baseDir string) (Storage, error) {
	// Expand home directory if needed
	if strings.HasPrefix(baseDir, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		baseDir = filepath.Join(home, baseDir[2:])
	}

	// Create base directory if it doesn't exist
	accountsDir := filepath.Join(baseDir, "accounts")
	if err := os.MkdirAll(accountsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create accounts directory: %w", err)
	}

	return &fileStorage{
		baseDir: baseDir,
	}, nil
}

// SaveSyncBuffer persists the get_updates_buf for resume capability
func (s *fileStorage) SaveSyncBuffer(ctx context.Context, accountID string, buffer []byte) error {
	filename := filepath.Join(s.baseDir, "accounts", fmt.Sprintf("%s.sync.json", accountID))

	// Write to temp file first, then rename (atomic operation)
	tempFile := filename + ".tmp"

	data := map[string]string{
		"buffer": string(buffer),
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal sync buffer: %w", err)
	}

	if err := os.WriteFile(tempFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := os.Rename(tempFile, filename); err != nil {
		os.Remove(tempFile) // Clean up temp file
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// LoadSyncBuffer retrieves the last saved sync buffer
func (s *fileStorage) LoadSyncBuffer(ctx context.Context, accountID string) ([]byte, error) {
	filename := filepath.Join(s.baseDir, "accounts", fmt.Sprintf("%s.sync.json", accountID))

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No sync buffer yet, return nil
		}
		return nil, fmt.Errorf("failed to read sync buffer file: %w", err)
	}

	var bufferData map[string]string
	if err := json.Unmarshal(data, &bufferData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal sync buffer: %w", err)
	}

	return []byte(bufferData["buffer"]), nil
}

// SaveAccount persists account credentials
func (s *fileStorage) SaveAccount(ctx context.Context, account *common.Account) error {
	filename := filepath.Join(s.baseDir, "accounts", fmt.Sprintf("%s.json", account.ID))

	// Write to temp file first, then rename (atomic operation)
	tempFile := filename + ".tmp"

	jsonData, err := json.MarshalIndent(account, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal account: %w", err)
	}

	if err := os.WriteFile(tempFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := os.Rename(tempFile, filename); err != nil {
		os.Remove(tempFile) // Clean up temp file
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	// Update accounts index
	if err := s.updateAccountsIndex(ctx, account.ID, true); err != nil {
		return fmt.Errorf("failed to update accounts index: %w", err)
	}

	return nil
}

// LoadAccount retrieves account by ID
func (s *fileStorage) LoadAccount(ctx context.Context, accountID string) (*common.Account, error) {
	filename := filepath.Join(s.baseDir, "accounts", fmt.Sprintf("%s.json", accountID))

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrAccountNotFound
		}
		return nil, fmt.Errorf("failed to read account file: %w", err)
	}

	var account common.Account
	if err := json.Unmarshal(data, &account); err != nil {
		return nil, fmt.Errorf("failed to unmarshal account: %w", err)
	}

	return &account, nil
}

// DeleteAccount removes account data
func (s *fileStorage) DeleteAccount(ctx context.Context, accountID string) error {
	// Delete account file
	accountFile := filepath.Join(s.baseDir, "accounts", fmt.Sprintf("%s.json", accountID))
	if err := os.Remove(accountFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete account file: %w", err)
	}

	// Delete sync buffer file
	syncFile := filepath.Join(s.baseDir, "accounts", fmt.Sprintf("%s.sync.json", accountID))
	if err := os.Remove(syncFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete sync buffer file: %w", err)
	}

	// Update accounts index
	if err := s.updateAccountsIndex(ctx, accountID, false); err != nil {
		return fmt.Errorf("failed to update accounts index: %w", err)
	}

	return nil
}

// ListAccounts returns all stored accounts
func (s *fileStorage) ListAccounts(ctx context.Context) ([]*common.Account, error) {
	indexFile := filepath.Join(s.baseDir, "accounts", "accounts.json")

	data, err := os.ReadFile(indexFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []*common.Account{}, nil // No accounts yet
		}
		return nil, fmt.Errorf("failed to read accounts index: %w", err)
	}

	var accountIDs []string
	if err := json.Unmarshal(data, &accountIDs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal accounts index: %w", err)
	}

	accounts := make([]*common.Account, 0, len(accountIDs))
	for _, id := range accountIDs {
		account, err := s.LoadAccount(ctx, id)
		if err != nil {
			if errors.Is(err, ErrAccountNotFound) {
				continue // Skip missing accounts
			}
			return nil, err
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

// updateAccountsIndex updates the accounts index file
func (s *fileStorage) updateAccountsIndex(ctx context.Context, accountID string, add bool) error {
	indexFile := filepath.Join(s.baseDir, "accounts", "accounts.json")

	// Read existing index
	var accountIDs []string
	data, err := os.ReadFile(indexFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read accounts index: %w", err)
	}

	if len(data) > 0 {
		if err := json.Unmarshal(data, &accountIDs); err != nil {
			return fmt.Errorf("failed to unmarshal accounts index: %w", err)
		}
	}

	// Update index
	if add {
		// Check if already exists
		found := false
		for _, id := range accountIDs {
			if id == accountID {
				found = true
				break
			}
		}
		if !found {
			accountIDs = append(accountIDs, accountID)
		}
	} else {
		// Remove from index
		filtered := make([]string, 0, len(accountIDs))
		for _, id := range accountIDs {
			if id != accountID {
				filtered = append(filtered, id)
			}
		}
		accountIDs = filtered
	}

	// Write updated index
	jsonData, err := json.MarshalIndent(accountIDs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal accounts index: %w", err)
	}

	tempFile := indexFile + ".tmp"
	if err := os.WriteFile(tempFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write temp index file: %w", err)
	}

	if err := os.Rename(tempFile, indexFile); err != nil {
		os.Remove(tempFile) // Clean up temp file
		return fmt.Errorf("failed to rename temp index file: %w", err)
	}

	return nil
}

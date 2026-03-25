package storage

import (
	"context"

	"github.com/waro163/wechat-bot-sdk/common"
)

// Storage abstracts persistence operations (file, DB, etc.)
type Storage interface {
	// SaveSyncBuffer persists the get_updates_buf for resume capability
	SaveSyncBuffer(ctx context.Context, accountID string, buffer []byte) error

	// LoadSyncBuffer retrieves the last saved sync buffer
	LoadSyncBuffer(ctx context.Context, accountID string) ([]byte, error)

	// SaveAccount persists account credentials
	SaveAccount(ctx context.Context, account *common.Account) error

	// LoadAccount retrieves account by ID
	LoadAccount(ctx context.Context, accountID string) (*common.Account, error)

	// DeleteAccount removes account data
	DeleteAccount(ctx context.Context, accountID string) error

	// ListAccounts returns all stored accounts
	ListAccounts(ctx context.Context) ([]*common.Account, error)
}

package wechatbotsdk

import (
	"fmt"
	"net/http"
	"time"

	"github.com/waro163/wechat-bot-sdk/api"
	"github.com/waro163/wechat-bot-sdk/cache"
	"github.com/waro163/wechat-bot-sdk/common"
	"github.com/waro163/wechat-bot-sdk/storage"
)

// Config holds the configuration for the WeChat bot SDK
type Config struct {
	// BaseURL is the API endpoint (e.g., https://ilinkai.weixin.qq.com)
	BaseURL string

	// CDNBaseURL for media uploads/downloads
	CDNBaseURL string

	// AccountID uniquely identifies this bot instance
	AccountID string

	// LongPollTimeout for getUpdates (default: 35s)
	LongPollTimeout time.Duration

	// APITimeout for regular API calls (default: 15s)
	APITimeout time.Duration

	// Storage implementation (file, DB, etc.)
	Storage storage.Storage

	// Cache implementation (memory, Redis, etc.)
	Cache cache.Cache

	// Logger implementation
	Logger common.Logger

	// HTTPClient for custom HTTP transport
	HTTPClient api.HTTPClient

	// MaxRetries for transient failures (default: 3)
	MaxRetries int

	// BackoffDelay for retry backoff (default: 30s)
	BackoffDelay time.Duration

	// RetryConfig for advanced retry configuration
	RetryConfig *RetryConfig

	// WorkerPoolSize for batch operations (default: 5)
	WorkerPoolSize int

	// MessageChannelBuffer is the buffer size for message channel (default: 100)
	MessageChannelBuffer int
}

// RetryConfig controls retry behavior
type RetryConfig struct {
	// MaxAttempts is the maximum number of retry attempts
	MaxAttempts int

	// InitialDelay is the initial delay before first retry
	InitialDelay time.Duration

	// MaxDelay is the maximum delay between retries
	MaxDelay time.Duration

	// Multiplier is the backoff multiplier (e.g., 2.0 for exponential backoff)
	Multiplier float64
}

// DefaultRetryConfig returns sensible defaults for retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:  common.DefaultMaxRetries,
		InitialDelay: common.DefaultRetryInitialDelay,
		MaxDelay:     common.DefaultRetryMaxDelay,
		Multiplier:   common.DefaultRetryMultiplier,
	}
}

// WithDefaults fills in zero values with sensible defaults
func (c *Config) WithDefaults() (*Config, error) {
	// Create a copy to avoid modifying the original
	cfg := *c

	if cfg.BaseURL == "" {
		cfg.BaseURL = common.DefaultBaseURL
	}

	if cfg.CDNBaseURL == "" {
		cfg.CDNBaseURL = common.DefaultCDNBaseURL
	}

	if cfg.LongPollTimeout == 0 {
		cfg.LongPollTimeout = common.DefaultLongPollTimeout
	}

	if cfg.APITimeout == 0 {
		cfg.APITimeout = common.DefaultAPITimeout
	}

	if cfg.Storage == nil {
		var err error
		cfg.Storage, err = storage.NewFileStorage(common.DefaultStateDir)
		if err != nil {
			return nil, fmt.Errorf("NewFileStorage error: %w", err)
		}
	}

	if cfg.Cache == nil {
		cfg.Cache = cache.NewMemoryCache()
	}

	if cfg.Logger == nil {
		cfg.Logger = common.NewDefaultLogger(cfg.AccountID)
	}

	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{
			Timeout: cfg.APITimeout,
		}
	}

	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = common.DefaultMaxRetries
	}

	if cfg.BackoffDelay == 0 {
		cfg.BackoffDelay = common.DefaultBackoffDelay
	}

	if cfg.RetryConfig == nil {
		cfg.RetryConfig = DefaultRetryConfig()
	}

	if cfg.WorkerPoolSize == 0 {
		cfg.WorkerPoolSize = common.DefaultWorkerPoolSize
	}

	if cfg.MessageChannelBuffer == 0 {
		cfg.MessageChannelBuffer = common.DefaultMessageChannelBuffer
	}

	return &cfg, nil
}

// Validate checks configuration correctness
func (c *Config) Validate() error {
	if c.BaseURL == "" {
		return common.ErrMissingBaseURL
	}

	if c.CDNBaseURL == "" {
		return common.ErrMissingCDNBaseURL
	}

	if c.AccountID == "" {
		return common.ErrMissingAccountID
	}

	if c.Storage == nil {
		return common.NewError(0, "Config.Validate", "storage implementation is required", common.ErrInvalidConfig)
	}

	if c.Cache == nil {
		return common.NewError(0, "Config.Validate", "cache implementation is required", common.ErrInvalidConfig)
	}

	if c.Logger == nil {
		return common.NewError(0, "Config.Validate", "logger implementation is required", common.ErrInvalidConfig)
	}

	if c.HTTPClient == nil {
		return common.NewError(0, "Config.Validate", "HTTP client is required", common.ErrInvalidConfig)
	}

	if c.LongPollTimeout <= 0 {
		return common.NewError(0, "Config.Validate", "long poll timeout must be positive", common.ErrInvalidConfig)
	}

	if c.APITimeout <= 0 {
		return common.NewError(0, "Config.Validate", "API timeout must be positive", common.ErrInvalidConfig)
	}

	if c.MaxRetries < 0 {
		return common.NewError(0, "Config.Validate", "max retries must be non-negative", common.ErrInvalidConfig)
	}

	if c.WorkerPoolSize <= 0 {
		return common.NewError(0, "Config.Validate", "worker pool size must be positive", common.ErrInvalidConfig)
	}

	if c.MessageChannelBuffer < 0 {
		return common.NewError(0, "Config.Validate", "message channel buffer must be non-negative", common.ErrInvalidConfig)
	}

	return nil
}

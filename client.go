package wechatbotsdk

import (
	"context"
	"fmt"
	"time"

	"github.com/waro163/wechat-bot-sdk/api"
	"github.com/waro163/wechat-bot-sdk/auth"
	"github.com/waro163/wechat-bot-sdk/cdn"
	"github.com/waro163/wechat-bot-sdk/common"
	"github.com/waro163/wechat-bot-sdk/messaging"
	"github.com/waro163/wechat-bot-sdk/monitor"
)

// Client is the main SDK entry point for WeChat bot operations
type Client struct {
	config     *Config
	apiClient  api.IClient
	uploader   *cdn.Uploader
	downloader *cdn.Downloader
	sender     *messaging.Sender
	monitor    *WrapMonitor
}

// New creates a new WeChat bot client
func New(cfg *Config) (*Client, error) {
	// Apply defaults
	var err error
	cfg, err = cfg.WithDefaults()
	if err != nil {
		return nil, fmt.Errorf("failed to apply defaults config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Create API client
	apiClient, err := api.NewClient(api.ClientOptions{
		BaseURL:    cfg.BaseURL,
		HTTPClient: cfg.HTTPClient,
		Logger:     cfg.Logger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create API client: %w", err)
	}

	// Create uploader
	uploader, err := cdn.NewUploader(
		cfg.CDNBaseURL,
		apiClient,
		cfg.Logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create CDN uploader: %w", err)
	}

	// Create downloader
	downloader, err := cdn.NewDownloader(
		cfg.CDNBaseURL,
		apiClient,
		cfg.Logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create CDN downloader: %w", err)
	}

	// Create sender (uses cache for context tokens)
	sender := messaging.NewSender(
		apiClient,
		cfg.Cache,
		cfg.Logger,
		cfg.WorkerPoolSize,
	)

	return &Client{
		config:     cfg,
		apiClient:  apiClient,
		uploader:   uploader,
		downloader: downloader,
		sender:     sender,
	}, nil
}

// Authenticate performs QR code authentication
func (c *Client) Authenticate(ctx context.Context, qrCodeDisplay func(string) error) (*common.AuthResult, error) {
	account, err := c.config.Storage.LoadAccount(ctx, c.config.AccountID)
	if err == nil {
		c.apiClient.SetBotToken(account.BotToken)
		c.apiClient.SetBaseURL(account.BaseURL)
		return &common.AuthResult{
			AccountID:   account.AccountID,
			BotToken:    account.BotToken,
			BaseURL:     account.BaseURL,
			ILinkBotID:  account.ILinkBotID,
			ILinkUserID: account.ILinkUserID,
		}, nil
	}
	authenticator := auth.NewQRAuthenticator(
		c.apiClient,
		c.config.Logger,
	)

	result, err := authenticator.Authenticate(ctx, c.config.AccountID, qrCodeDisplay)
	if err != nil {
		return nil, err
	}

	c.apiClient.SetBotToken(result.BotToken)
	c.apiClient.SetBaseURL(result.BaseURL)

	// Save account to storage
	account = &common.Account{
		AuthResult: *result,
		CreatedAt:  time.Now().Unix(),
		UpdatedAt:  time.Now().Unix(),
	}

	if err := c.config.Storage.SaveAccount(ctx, account); err != nil {
		c.config.Logger.Warn("Failed to save account", common.Field{Key: "error", Value: err.Error()})
	}

	// Convert to public type
	return &common.AuthResult{
		AccountID:   result.AccountID,
		BotToken:    result.BotToken,
		BaseURL:     result.BaseURL,
		ILinkBotID:  result.ILinkBotID,
		ILinkUserID: result.ILinkUserID,
	}, nil
}

// SendTextMessage sends a text message to a user
func (c *Client) SendTextMessage(ctx context.Context, toUserID, text string) error {
	return c.sender.SendText(ctx, toUserID, text)
}

// uploadMedia uploads a media file and returns upload info
func (c *Client) uploadMedia(ctx context.Context, fileData []byte, toUserID string, mediaType common.UploadMediaType) (*common.UploadedFileInfo, error) {
	info, err := c.uploader.UploadFile(
		ctx,
		fileData,
		toUserID,
		mediaType,
	)
	if err != nil {
		return nil, err
	}
	return info, nil
}

// SendImageMessage sends an image message
func (c *Client) SendImageMessage(ctx context.Context, toUserID string, image []byte) error {
	// Upload image
	media, err := c.uploadMedia(ctx, image, toUserID, common.UploadMediaTypeImage)
	if err != nil {
		return fmt.Errorf("failed to upload image: %w", err)
	}

	// Create CDN media reference
	encryptType := common.EncryptTypePackaged
	cdnMedia := &common.CDNMedia{
		EncryptQueryParam: &media.DownloadEncryptedQueryParam,
		AESKey:            &media.AESKeyHex,
		EncryptType:       &encryptType,
	}

	// Send message
	return c.sender.SendImage(ctx, toUserID, cdnMedia, media.FileSizeCiphertext)
}

// SendFileMessage sends a file message
func (c *Client) SendFileMessage(ctx context.Context, toUserID, fileName string, file []byte) error {
	// Upload file
	media, err := c.uploadMedia(ctx, file, toUserID, common.UploadMediaTypeFile)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	// Create CDN media reference
	encryptType := common.EncryptTypePackaged
	cdnMedia := &common.CDNMedia{
		EncryptQueryParam: &media.DownloadEncryptedQueryParam,
		AESKey:            &media.AESKeyHex,
		EncryptType:       &encryptType,
	}

	// Send message
	return c.sender.SendFile(ctx, toUserID, fileName, cdnMedia)
}

// SendVideoMessage sends a video message
func (c *Client) SendVideoMessage(ctx context.Context, toUserID string, video []byte) error {
	// Upload video
	media, err := c.uploadMedia(ctx, video, toUserID, common.UploadMediaTypeVideo)
	if err != nil {
		return fmt.Errorf("failed to upload video: %w", err)
	}

	// Create CDN media reference
	encryptType := common.EncryptTypePackaged
	cdnMedia := &common.CDNMedia{
		EncryptQueryParam: &media.DownloadEncryptedQueryParam,
		AESKey:            &media.AESKeyHex,
		EncryptType:       &encryptType,
	}

	// Send message
	return c.sender.SendVideo(ctx, toUserID, cdnMedia, media.FileSizeCiphertext)
}

// DownloadMedia downloads and decrypts media from a message
func (c *Client) DownloadMedia(ctx context.Context, encryptedQueryParam, aesKey string) ([]byte, error) {
	return c.downloader.DownloadAndDecrypt(ctx, encryptedQueryParam, aesKey)
}

// StartMonitor begins receiving messages
func (c *Client) StartMonitor() (*WrapMonitor, error) {
	if c.monitor != nil {
		return nil, common.ErrMonitorAlreadyStarted
	}

	// Create monitor
	mon := monitor.NewMonitor(monitor.MonitorConfig{
		AccountID:  c.config.AccountID,
		HttpClient: c.apiClient,
		Storage:    c.config.Storage,
		Logger:     c.config.Logger,
		Timeout:    c.config.LongPollTimeout,
		BufSize:    c.config.MessageChannelBuffer,
	})

	if err := mon.Start(); err != nil {
		return nil, err
	}

	// Create wrapper
	wrapper := &WrapMonitor{
		msgChan:  make(chan *common.Message, c.config.MessageChannelBuffer),
		internal: mon,
		sender:   c.sender,
	}

	c.monitor = wrapper

	return wrapper, nil
}

// Close gracefully shuts down the client
func (c *Client) Close(timeout time.Duration) error {
	if c.monitor != nil {
		return c.monitor.Stop(timeout)
	}
	return nil
}

// WrapMonitor wraps the internal monitor and provides message handling
type WrapMonitor struct {
	msgChan  chan *common.Message
	internal *monitor.Monitor
	sender   *messaging.Sender
}

// Stop gracefully stops the monitor
func (m *WrapMonitor) Stop(timeout time.Duration) error {
	close(m.msgChan)
	return m.internal.Stop(timeout)
}

// Errors returns the channel for receiving errors
func (m *WrapMonitor) Errors() <-chan error {
	return m.internal.Errors()
}

// Messages returns the channel for receiving messages
func (m *WrapMonitor) Messages() <-chan *common.Message {
	go func() {
		for msg := range m.internal.Messages() {
			// Store context token for replies
			if msg.ContextToken != "" {
				ctx := context.Background()
				m.sender.SetContextToken(ctx, msg.FromUserID, msg.ContextToken)
			}
			m.msgChan <- msg
		}
	}()

	return m.msgChan
}

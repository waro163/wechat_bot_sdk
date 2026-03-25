package messaging

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"

	"github.com/waro163/wechat-bot-sdk/api"
	"github.com/waro163/wechat-bot-sdk/cache"
	"github.com/waro163/wechat-bot-sdk/common"
)

// Sender handles message sending operations
type Sender struct {
	httpClient        api.IMessageClient
	contextTokenCache cache.Cache
	logger            common.Logger
	workerPoolSize    int
}

// NewSender creates a new message sender
func NewSender(httpClient api.IMessageClient, contextTokenCache cache.Cache, logger common.Logger, workerPoolSize int) *Sender {
	if workerPoolSize <= 0 {
		workerPoolSize = common.DefaultWorkerPoolSize
	}

	return &Sender{
		httpClient:        httpClient,
		contextTokenCache: contextTokenCache,
		logger:            logger,
		workerPoolSize:    workerPoolSize,
	}
}

// generateClientID generates a random client ID
func generateClientID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// getContextToken retrieves the context token for a user
func (s *Sender) getContextToken(ctx context.Context, toUserID string) (string, error) {
	value, err := s.contextTokenCache.Get(ctx, toUserID)
	if err != nil {
		return "", err
	}
	return string(value), nil
}

// SetContextToken stores a context token for a user
func (s *Sender) SetContextToken(ctx context.Context, fromUserID, contextToken string) error {
	return s.contextTokenCache.Set(ctx, fromUserID, []byte(contextToken), 0) // No expiration
}

// SendText sends a text message
func (s *Sender) SendText(ctx context.Context, toUserID, text string) error {
	// Get context token
	contextToken, err := s.getContextToken(ctx, toUserID)
	if err != nil {
		s.logger.Warn("Failed to get context token", common.Field{Key: "error", Value: err.Error()})
		contextToken = "" // Continue without context token
	}

	// Generate client ID
	clientID := generateClientID()

	// Build message
	msgType := common.MessageTypeBot
	msgState := common.MessageStateFinish
	itemType := common.MessageItemTypeText

	msg := &common.WeixinMessage{
		ToUserID:     &toUserID,
		ClientID:     &clientID,
		MessageType:  &msgType,
		MessageState: &msgState,
		ContextToken: &contextToken,
		ItemList: []common.MessageItem{
			{
				Type: &itemType,
				TextItem: &common.TextItem{
					Text: &text,
				},
			},
		},
	}

	req := &common.SendMessageRequest{Msg: msg}

	s.logger.Debug("Sending text message",
		common.Field{Key: "to", Value: toUserID},
		common.Field{Key: "textLen", Value: len(text)},
	)

	return s.httpClient.SendMessage(ctx, req)
}

// SendImage sends an image message
func (s *Sender) SendImage(ctx context.Context, toUserID string, media *common.CDNMedia, fileSize int64) error {
	contextToken, err := s.getContextToken(ctx, toUserID)
	if err != nil {
		s.logger.Warn("Failed to get context token", common.Field{Key: "error", Value: err.Error()})
		contextToken = ""
	}

	clientID := generateClientID()
	msgType := common.MessageTypeBot
	msgState := common.MessageStateFinish
	itemType := common.MessageItemTypeImage

	msg := &common.WeixinMessage{
		ToUserID:     &toUserID,
		ClientID:     &clientID,
		MessageType:  &msgType,
		MessageState: &msgState,
		ContextToken: &contextToken,
		ItemList: []common.MessageItem{
			{
				Type: &itemType,
				ImageItem: &common.ImageItem{
					Media:   media,
					MidSize: &fileSize,
				},
			},
		},
	}

	req := &common.SendMessageRequest{Msg: msg}

	s.logger.Debug("Sending image message", common.Field{Key: "to", Value: toUserID})

	return s.httpClient.SendMessage(ctx, req)
}

// SendFile sends a file message
func (s *Sender) SendFile(ctx context.Context, toUserID, fileName string, media *common.CDNMedia) error {
	contextToken, err := s.getContextToken(ctx, toUserID)
	if err != nil {
		s.logger.Warn("Failed to get context token", common.Field{Key: "error", Value: err.Error()})
		contextToken = ""
	}

	clientID := generateClientID()
	msgType := common.MessageTypeBot
	msgState := common.MessageStateFinish
	itemType := common.MessageItemTypeFile

	msg := &common.WeixinMessage{
		ToUserID:     &toUserID,
		ClientID:     &clientID,
		MessageType:  &msgType,
		MessageState: &msgState,
		ContextToken: &contextToken,
		ItemList: []common.MessageItem{
			{
				Type: &itemType,
				FileItem: &common.FileItem{
					Media:    media,
					FileName: &fileName,
				},
			},
		},
	}

	req := &common.SendMessageRequest{Msg: msg}

	s.logger.Debug("Sending file message",
		common.Field{Key: "to", Value: toUserID},
		common.Field{Key: "fileName", Value: fileName},
	)

	return s.httpClient.SendMessage(ctx, req)
}

// SendVideo sends a video message
func (s *Sender) SendVideo(ctx context.Context, toUserID string, media *common.CDNMedia, fileSize int64) error {
	contextToken, err := s.getContextToken(ctx, toUserID)
	if err != nil {
		s.logger.Warn("Failed to get context token", common.Field{Key: "error", Value: err.Error()})
		contextToken = ""
	}

	clientID := generateClientID()
	msgType := common.MessageTypeBot
	msgState := common.MessageStateFinish
	itemType := common.MessageItemTypeVideo

	msg := &common.WeixinMessage{
		ToUserID:     &toUserID,
		ClientID:     &clientID,
		MessageType:  &msgType,
		MessageState: &msgState,
		ContextToken: &contextToken,
		ItemList: []common.MessageItem{
			{
				Type: &itemType,
				VideoItem: &common.VideoItem{
					Media:     media,
					VideoSize: &fileSize,
				},
			},
		},
	}

	req := &common.SendMessageRequest{Msg: msg}

	s.logger.Debug("Sending video message", common.Field{Key: "to", Value: toUserID})

	return s.httpClient.SendMessage(ctx, req)
}

type batchResult struct {
	index int
	err   error
}

// BatchSendText sends multiple text messages concurrently using a worker pool
func (s *Sender) BatchSendText(ctx context.Context, messages []common.TextMessage) []error {
	jobsChan := make(chan common.TextMessage, len(messages))
	resultsChan := make(chan batchResult, len(messages))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < s.workerPoolSize; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for msg := range jobsChan {
				err := s.SendText(ctx, msg.ToUserID, msg.Text)
				resultsChan <- batchResult{
					index: msg.Index,
					err:   err,
				}
			}
		}()
	}

	// Send jobs
	for _, msg := range messages {
		jobsChan <- msg
	}
	close(jobsChan)

	// Wait for completion
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	results := make([]error, len(messages))
	for result := range resultsChan {
		if result.index < len(results) {
			results[result.index] = result.err
		}
	}

	return results
}

package monitor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/waro163/wechat-bot-sdk/api"
	"github.com/waro163/wechat-bot-sdk/common"
	"github.com/waro163/wechat-bot-sdk/storage"
)

// MonitorConfig holds configuration for the monitor
type MonitorConfig struct {
	AccountID  string
	HttpClient api.IMonitorClient
	Storage    storage.Storage
	Logger     common.Logger
	Timeout    time.Duration
	BufSize    int // Message channel buffer size
}

// Monitor manages the long-poll message receiving loop
type Monitor struct {
	accountID   string
	httpClient  api.IMonitorClient
	storage     storage.Storage
	msgChan     chan *common.Message
	errChan     chan error
	stopChan    chan struct{}
	stoppedChan chan struct{}
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	logger      common.Logger
	timeout     time.Duration
}

// NewMonitor creates a new message monitor
func NewMonitor(cfg MonitorConfig) *Monitor {
	if cfg.Timeout == 0 {
		cfg.Timeout = common.DefaultLongPollTimeout
	}
	if cfg.BufSize == 0 {
		cfg.BufSize = common.DefaultMessageChannelBuffer
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Monitor{
		accountID:   cfg.AccountID,
		httpClient:  cfg.HttpClient,
		storage:     cfg.Storage,
		msgChan:     make(chan *common.Message, cfg.BufSize),
		errChan:     make(chan error, 10),
		stopChan:    make(chan struct{}),
		stoppedChan: make(chan struct{}),
		ctx:         ctx,
		cancel:      cancel,
		logger:      cfg.Logger,
		timeout:     cfg.Timeout,
	}
}

// Start begins the long-poll loop in a background goroutine
func (m *Monitor) Start() error {
	m.logger.Info("Starting monitor", common.Field{Key: "accountID", Value: m.accountID})

	m.wg.Add(1)
	go m.pollLoop()

	return nil
}

// Messages returns the channel for receiving inbound messages
func (m *Monitor) Messages() <-chan *common.Message {
	return m.msgChan
}

// Errors returns the channel for receiving errors
func (m *Monitor) Errors() <-chan error {
	return m.errChan
}

// Stop gracefully shuts down the monitor
func (m *Monitor) Stop(timeout time.Duration) error {
	m.logger.Info("Stopping monitor", common.Field{Key: "accountID", Value: m.accountID})

	// Signal stop
	m.cancel()
	close(m.stopChan)

	// Wait for goroutines to finish
	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		close(m.msgChan)
		close(m.errChan)
		m.logger.Info("Monitor stopped gracefully")
		return nil
	case <-time.After(timeout):
		m.logger.Error("Monitor stop timeout exceeded")
		return fmt.Errorf("stop timeout exceeded")
	}
}

// pollLoop is the main long-polling loop
func (m *Monitor) pollLoop() {
	defer m.wg.Done()

	// Load previous sync buffer
	syncBuf, err := m.storage.LoadSyncBuffer(m.ctx, m.accountID)
	if err != nil {
		m.logger.Warn("Failed to load sync buffer, starting fresh", common.Field{Key: "error", Value: err.Error()})
		syncBuf = nil
	}

	syncBufStr := string(syncBuf)
	if syncBufStr != "" {
		m.logger.Info("Resuming from previous sync buffer", common.Field{Key: "syncBufStr", Value: syncBufStr})
	} else {
		m.logger.Info("No previous sync buffer, starting fresh")
	}

	consecutiveFailures := 0
	nextTimeout := m.timeout

	for {
		// Check if stopped
		select {
		case <-m.ctx.Done():
			m.logger.Info("Monitor context canceled, stopping poll loop")
			return
		case <-m.stopChan:
			m.logger.Info("Monitor stop signal received, stopping poll loop")
			return
		default:
		}

		// Call getUpdates
		m.logger.Debug("Calling getUpdates",
			common.Field{Key: "timeout", Value: nextTimeout.String()},
			common.Field{Key: "syncBufStr", Value: syncBufStr},
		)

		req := &common.GetUpdatesRequest{
			GetUpdatesBuf: syncBufStr,
		}

		resp, err := m.httpClient.GetUpdates(m.ctx, req, nextTimeout)
		if err != nil {
			consecutiveFailures++
			m.logger.Error("GetUpdates failed",
				common.Field{Key: "error", Value: err.Error()},
				common.Field{Key: "failures", Value: consecutiveFailures},
			)

			// Send error to channel (non-blocking)
			select {
			case m.errChan <- err:
			default:
			}

			// Backoff after consecutive failures
			if consecutiveFailures >= common.MaxConsecutiveFailures {
				m.logger.Warn("Max consecutive failures reached, backing off",
					common.Field{Key: "delay", Value: common.DefaultBackoffDelay.String()},
				)
				consecutiveFailures = 0
				m.sleep(common.DefaultBackoffDelay)
			} else {
				m.sleep(common.DefaultRetryInitialDelay)
			}
			continue
		}

		// Check for API errors
		if resp.ErrCode != 0 || resp.Ret != 0 {
			// Check for session expiration
			if resp.ErrCode == common.SessionExpiredErrorCode || resp.Ret == common.SessionExpiredErrorCode {
				m.logger.Error("Session expired, pausing",
					common.Field{Key: "pauseDuration", Value: common.SessionPauseDuration.String()},
				)

				// Send error to channel
				select {
				case m.errChan <- fmt.Errorf("session expired (code=%d)", common.SessionExpiredErrorCode):
				default:
				}

				consecutiveFailures = 0
				m.sleep(common.SessionPauseDuration)
				continue
			}

			// Other API errors
			consecutiveFailures++
			m.logger.Error("API error",
				common.Field{Key: "errCode", Value: resp.ErrCode},
				common.Field{Key: "ret", Value: resp.Ret},
				common.Field{Key: "errMsg", Value: resp.ErrMsg},
			)

			if consecutiveFailures >= common.MaxConsecutiveFailures {
				m.logger.Warn("Max consecutive failures reached, backing off")
				consecutiveFailures = 0
				m.sleep(common.DefaultBackoffDelay)
			} else {
				m.sleep(common.DefaultRetryInitialDelay)
			}
			continue
		}

		// Success - reset failure counter
		consecutiveFailures = 0

		// Update timeout if server suggests one
		if resp.LongPollingTimeoutMs > 0 {
			nextTimeout = time.Duration(resp.LongPollingTimeoutMs) * time.Millisecond
			m.logger.Debug("Updated poll timeout", common.Field{Key: "timeout", Value: nextTimeout.String()})
		}

		// Save sync buffer
		if resp.GetUpdatesBuf != "" {
			syncBufStr = resp.GetUpdatesBuf
			if err := m.storage.SaveSyncBuffer(m.ctx, m.accountID, []byte(syncBufStr)); err != nil {
				m.logger.Warn("Failed to save sync buffer",
					common.Field{Key: "error", Value: err.Error()},
					common.Field{Key: "syncBufStr", Value: syncBufStr},
				)
			}
		}

		// Process messages
		for _, msgData := range resp.Msgs {
			msg := m.convertMessage(&msgData)
			if msg != nil {
				// Send message to channel (non-blocking)
				select {
				case m.msgChan <- msg:
					m.logger.Debug("Message sent to channel", common.Field{Key: "from", Value: msg.FromUserID})
				case <-m.ctx.Done():
					return
				default:
					m.logger.Warn("Message channel full, dropping message")
				}
			}
		}
	}
}

// sleep sleeps for the given duration, checking for stop signal
func (m *Monitor) sleep(d time.Duration) {
	select {
	case <-time.After(d):
	case <-m.ctx.Done():
	case <-m.stopChan:
	}
}

// convertMessage converts API message data to internal Message type
func (m *Monitor) convertMessage(data *common.WeixinMessage) *common.Message {
	if data.FromUserID == nil {
		return nil
	}

	msg := &common.Message{
		FromUserID: *data.FromUserID,
		Items:      data.ItemList,
	}

	if data.ToUserID != nil {
		msg.ToUserID = *data.ToUserID
	}
	if data.ContextToken != nil {
		msg.ContextToken = *data.ContextToken
	}
	if data.CreateTimeMs != nil {
		msg.CreateTimeMs = *data.CreateTimeMs
	}

	return msg
}

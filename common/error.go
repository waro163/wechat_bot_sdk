package common

import (
	"errors"
	"fmt"
	"net"
)

// ErrorCode represents API error codes
type ErrorCode int

const (
	ErrCodeSessionExpired ErrorCode = -14
	ErrCodeInvalidToken   ErrorCode = -1
	ErrCodeRateLimited    ErrorCode = -429
)

// Sentinel errors for common cases
var (
	// Configuration errors
	ErrMissingBaseURL    = errors.New("missing base URL in config")
	ErrMissingCDNBaseURL = errors.New("missing CDN base URL in config")
	ErrMissingAccountID  = errors.New("missing account ID in config")
	ErrMissingToken      = errors.New("missing authentication token")
	ErrInvalidConfig     = errors.New("invalid configuration")

	// Authentication errors
	ErrSessionExpired       = errors.New("session expired, re-authentication required")
	ErrQRCodeExpired        = errors.New("QR code expired")
	ErrQRCodeScanTimeout    = errors.New("QR code scan timeout")
	ErrAuthenticationFailed = errors.New("authentication failed")
	ErrInvalidToken         = errors.New("invalid authentication token")

	// Network errors
	ErrNetworkTimeout   = errors.New("network timeout")
	ErrConnectionFailed = errors.New("connection failed")
	ErrRequestFailed    = errors.New("request failed")

	// Operation errors
	ErrStopTimeout           = errors.New("stop timeout exceeded")
	ErrContextCanceled       = errors.New("context canceled")
	ErrMonitorNotStarted     = errors.New("monitor not started")
	ErrMonitorAlreadyStarted = errors.New("monitor already started")

	// Storage errors
	ErrAccountNotFound        = errors.New("account not found")
	ErrAccountAlreadyExists   = errors.New("account already exists")
	ErrStorageOperationFailed = errors.New("storage operation failed")
	ErrCacheOperationFailed   = errors.New("cache operation failed")

	// Media errors
	ErrMediaUploadFailed    = errors.New("media upload failed")
	ErrMediaDownloadFailed  = errors.New("media download failed")
	ErrEncryptionFailed     = errors.New("encryption failed")
	ErrDecryptionFailed     = errors.New("decryption failed")
	ErrUnsupportedMediaType = errors.New("unsupported media type")
	ErrInvalidMediaFile     = errors.New("invalid media file")

	// Message errors
	ErrSendMessageFailed    = errors.New("send message failed")
	ErrReceiveMessageFailed = errors.New("receive message failed")
	ErrInvalidMessage       = errors.New("invalid message")
	ErrMissingContextToken  = errors.New("missing context token")
	ErrEmptyMessageContent  = errors.New("empty message content")

	// API errors
	ErrAPICallFailed      = errors.New("API call failed")
	ErrInvalidAPIResponse = errors.New("invalid API response")
	ErrSessionPaused      = errors.New("session paused")
)

// Error wraps SDK errors with context
type Error struct {
	Code    ErrorCode
	Message string
	Cause   error
	Op      string // Operation that failed
	Account string // AccountID if relevant
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s failed: %s (code=%d): %v", e.Op, e.Message, e.Code, e.Cause)
	}
	return fmt.Sprintf("%s failed: %s (code=%d)", e.Op, e.Message, e.Code)
}

// Unwrap implements error unwrapping for errors.Is/As
func (e *Error) Unwrap() error {
	return e.Cause
}

// NewError creates a new Error with the given parameters
func NewError(code ErrorCode, op, message string, cause error) *Error {
	return &Error{
		Code:    code,
		Op:      op,
		Message: message,
		Cause:   cause,
	}
}

// NewErrorWithAccount creates a new Error with account context
func NewErrorWithAccount(code ErrorCode, op, message, accountID string, cause error) *Error {
	return &Error{
		Code:    code,
		Op:      op,
		Message: message,
		Account: accountID,
		Cause:   cause,
	}
}

// IsRetryable determines if an error should trigger a retry
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Check for custom SDK errors
	var wErr *Error
	if errors.As(err, &wErr) {
		switch wErr.Code {
		case ErrCodeSessionExpired:
			return false // Needs re-auth, not just retry
		case ErrCodeInvalidToken:
			return false // Needs re-auth
		case ErrCodeRateLimited:
			return true // Rate limit, should retry with backoff
		default:
			return false
		}
	}

	// Check for specific sentinel errors
	switch {
	case errors.Is(err, ErrSessionExpired):
		return false
	case errors.Is(err, ErrAuthenticationFailed):
		return false
	case errors.Is(err, ErrInvalidToken):
		return false
	case errors.Is(err, ErrQRCodeExpired):
		return false
	case errors.Is(err, ErrContextCanceled):
		return false
	}

	// Check for network errors that are generally retryable
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	// Default: not retryable
	return false
}

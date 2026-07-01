package common

import "time"

const (
	// DefaultBaseURL is the default WeChat API base URL
	DefaultBaseURL = "https://ilinkai.weixin.qq.com"

	// DefaultCDNBaseURL is the default CDN base URL for media uploads/downloads
	DefaultCDNBaseURL = "https://novac2c.cdn.weixin.qq.com/c2c"

	// DefaultLongPollTimeout is the default timeout for long-polling getUpdates
	DefaultLongPollTimeout = 35 * time.Second

	// DefaultAPITimeout is the default timeout for regular API calls
	DefaultAPITimeout = 60 * time.Second

	// DefaultConfigTimeout is the default timeout for config fetching
	DefaultConfigTimeout = 10 * time.Second

	// DefaultQRPollTimeout is the default timeout for QR code status polling
	DefaultQRPollTimeout = 35 * time.Second

	// DefaultMaxRetries is the default number of retry attempts for transient failures
	DefaultMaxRetries = 3

	// DefaultBackoffDelay is the default delay after consecutive failures
	DefaultBackoffDelay = 30 * time.Second

	// DefaultRetryInitialDelay is the initial delay for exponential backoff
	DefaultRetryInitialDelay = 2 * time.Second

	// DefaultRetryMaxDelay is the maximum delay for exponential backoff
	DefaultRetryMaxDelay = 30 * time.Second

	// DefaultRetryMultiplier is the multiplier for exponential backoff
	DefaultRetryMultiplier = 2.0

	// DefaultWorkerPoolSize is the default number of concurrent workers for batch operations
	DefaultWorkerPoolSize = 5

	// DefaultMessageChannelBuffer is the buffer size for the message channel
	DefaultMessageChannelBuffer = 100

	// MaxConsecutiveFailures before backoff
	MaxConsecutiveFailures = 3

	// SessionExpiredErrorCode is the API error code for session expiration
	SessionExpiredErrorCode = -14

	// SessionPauseDuration is the duration to pause after session expiration
	SessionPauseDuration = 60 * time.Minute

	// ConfigCacheTTL is the TTL for config cache entries
	ConfigCacheTTL = 24 * time.Hour

	// ContextTokenCacheTTL is the TTL for context token cache (0 = no expiration)
	ContextTokenCacheTTL = 0

	// QRCodeMaxRefreshCount is the maximum number of QR code refresh attempts
	QRCodeMaxRefreshCount = 3

	// QRCodeDefaultTTL is the default TTL for QR codes
	QRCodeDefaultTTL = 5 * time.Minute

	// UploadMaxRetries is the maximum number of retry attempts for CDN uploads
	UploadMaxRetries = 3

	// MaxConcurrentUploads is the maximum number of concurrent media uploads
	MaxConcurrentUploads = 5
)

const (
	// HTTP header names
	HeaderContentType   = "Content-Type"
	HeaderAuthType      = "AuthorizationType"
	HeaderAuthorization = "Authorization"
	HeaderWechatUIN     = "X-WECHAT-UIN"
	HeaderRouteTag      = "SKRouteTag"

	// Auth type value
	AuthTypeILinkBotToken = "ilink_bot_token"

	ILinkAppId            = "bot"
	ILinkAppClientVersion = "0x00020404"

	// Content types
	ContentTypeJSON        = "application/json"
	ContentTypeOctetStream = "application/octet-stream"

	// ChannelVersion
	ChannelVersion = "1.0.0"
)

const (
	// API endpoints (all relative to base URL)
	EndpointGetUpdates   = "/ilink/bot/getupdates"
	EndpointSendMessage  = "/ilink/bot/sendmessage"
	EndpointGetUploadURL = "/ilink/bot/getuploadurl"
	EndpointGetConfig    = "/ilink/bot/getconfig"
	EndpointSendTyping   = "/ilink/bot/sendtyping"
	EndpointGetBotQRCode = "/ilink/bot/get_bot_qrcode"
	EndpointGetQRStatus  = "/ilink/bot/get_qrcode_status"
)

const (
	// DefaultBotType is the default bot type for QR code authentication
	DefaultBotType = "3"

	// QRCodeTTL is the TTL for QR codes
	QRCodeTTL = 5 * time.Minute

	// QRLongPollTimeout is the timeout for QR status polling
	QRLongPollTimeout = 35 * time.Second

	// MaxQRRefreshCount is the maximum number of QR refresh attempts
	MaxQRRefreshCount = 3
)

// QRStatus represents the status of a QR code
type QRStatus string

const (
	QRStatusWait              QRStatus = "wait"
	QRStatusScanned           QRStatus = "scaned" // Note: typo from API
	QRStatusConfirmed         QRStatus = "confirmed"
	QRStatusExpired           QRStatus = "expired"
	QRStatusNeedVerifyCode    QRStatus = "need_verifycode"
	QRStatusVerifyCodeBlocked QRStatus = "verify_code_blocked"
	QRStatusBindedRedirect    QRStatus = "binded_redirect"
	QRStatusScanedButRedirect QRStatus = "scaned_but_redirect"
)

const (
	// Typing status values
	TypingStatusTyping = 1
	TypingStatusCancel = 2
)

const (
	// Voice encoding types
	VoiceEncodePCM      = 1
	VoiceEncodeADPCM    = 2
	VoiceEncodeFeature  = 3
	VoiceEncodeSpeex    = 4
	VoiceEncodeAMR      = 5
	VoiceEncodeSilk     = 6
	VoiceEncodeMP3      = 7
	VoiceEncodeOggSpeex = 8
)

const (
	// Default storage paths
	DefaultStateDir    = "./.wechat-bot"
	AccountsSubDir     = "accounts"
	AccountFilePattern = "%s.json"      // accountID.json
	SyncBufFilePattern = "%s.sync.json" // accountID.sync.json
	AccountsIndexFile  = "accounts.json"
)

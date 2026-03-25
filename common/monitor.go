package common

// Message represents a WeChat message (simplified for monitor)
type Message struct {
	FromUserID   string
	ToUserID     string
	ContextToken string
	Items        []MessageItem
	CreateTimeMs int64
}

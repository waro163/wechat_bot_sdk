package common

// MessageType represents the type of message (user or bot)
type MessageType int

const (
	MessageTypeNone MessageType = 0
	MessageTypeUser MessageType = 1
	MessageTypeBot  MessageType = 2
)

// MessageItemType represents the type of content in a message item
type MessageItemType int

const (
	MessageItemTypeNone  MessageItemType = 0
	MessageItemTypeText  MessageItemType = 1
	MessageItemTypeImage MessageItemType = 2
	MessageItemTypeVoice MessageItemType = 3
	MessageItemTypeFile  MessageItemType = 4
	MessageItemTypeVideo MessageItemType = 5
)

// MessageState represents the state of a message
type MessageState int

const (
	MessageStateNew        MessageState = 0
	MessageStateGenerating MessageState = 1
	MessageStateFinish     MessageState = 2
)

// EncryptType represents the CDN encryption type
type EncryptType int

const (
	EncryptTypeFileIDOnly EncryptType = 0 // Only encrypt file ID
	EncryptTypePackaged   EncryptType = 1 // Package thumbnail/mid-size info
)

package common

// WeixinMessage represents a WeChat message (internal API representation)
type WeixinMessage struct {
	Seq          *int64        `json:"seq,omitempty"`
	MessageID    *int64        `json:"message_id,omitempty"`
	FromUserID   *string       `json:"from_user_id,omitempty"`
	ToUserID     *string       `json:"to_user_id,omitempty"`
	ClientID     *string       `json:"client_id,omitempty"`
	CreateTimeMs *int64        `json:"create_time_ms,omitempty"`
	UpdateTimeMs *int64        `json:"update_time_ms,omitempty"`
	DeleteTimeMs *int64        `json:"delete_time_ms,omitempty"`
	SessionID    *string       `json:"session_id,omitempty"`
	GroupID      *string       `json:"group_id,omitempty"`
	MessageType  *MessageType  `json:"message_type,omitempty"`
	MessageState *MessageState `json:"message_state,omitempty"`
	ItemList     []MessageItem `json:"item_list,omitempty"`
	ContextToken *string       `json:"context_token,omitempty"`
}

// MessageItem represents a single content item in a message
type MessageItem struct {
	Type         *MessageItemType `json:"type,omitempty"`
	CreateTimeMs *int64           `json:"create_time_ms,omitempty"`
	UpdateTimeMs *int64           `json:"update_time_ms,omitempty"`
	IsCompleted  *bool            `json:"is_completed,omitempty"`
	MsgID        *string          `json:"msg_id,omitempty"`
	RefMsg       *RefMessage      `json:"ref_msg,omitempty"`
	TextItem     *TextItem        `json:"text_item,omitempty"`
	ImageItem    *ImageItem       `json:"image_item,omitempty"`
	VoiceItem    *VoiceItem       `json:"voice_item,omitempty"`
	FileItem     *FileItem        `json:"file_item,omitempty"`
	VideoItem    *VideoItem       `json:"video_item,omitempty"`
}

// TextItem represents a text message content
type TextItem struct {
	Text *string `json:"text,omitempty"`
}

// RefMessage represents a referenced/quoted message
type RefMessage struct {
	MessageItem *MessageItem `json:"message_item,omitempty"`
	Title       *string      `json:"title,omitempty"`
}

// CDNMedia represents a CDN media reference with encryption info
type CDNMedia struct {
	FullURL           *string      `json:"full_url,omitempty"`
	EncryptQueryParam *string      `json:"encrypt_query_param,omitempty"`
	AESKey            *string      `json:"aes_key,omitempty"`
	EncryptType       *EncryptType `json:"encrypt_type,omitempty"`
}

// ImageItem represents an image message content
type ImageItem struct {
	Media       *CDNMedia `json:"media,omitempty"`
	ThumbMedia  *CDNMedia `json:"thumb_media,omitempty"`
	AESKey      *string   `json:"aeskey,omitempty"`
	URL         *string   `json:"url,omitempty"`
	MidSize     *int64    `json:"mid_size,omitempty"`
	ThumbSize   *int64    `json:"thumb_size,omitempty"`
	ThumbHeight *int      `json:"thumb_height,omitempty"`
	ThumbWidth  *int      `json:"thumb_width,omitempty"`
	HDSize      *int64    `json:"hd_size,omitempty"`
}

// VoiceItem represents a voice/audio message content
type VoiceItem struct {
	Media         *CDNMedia `json:"media,omitempty"`
	EncodeType    *int      `json:"encode_type,omitempty"`
	BitsPerSample *int      `json:"bits_per_sample,omitempty"`
	SampleRate    *int      `json:"sample_rate,omitempty"`
	PlayTime      *int      `json:"playtime,omitempty"`
	Text          *string   `json:"text,omitempty"`
}

// FileItem represents a file attachment message content
type FileItem struct {
	Media    *CDNMedia `json:"media,omitempty"`
	FileName *string   `json:"file_name,omitempty"`
	MD5      *string   `json:"md5,omitempty"`
	Len      *string   `json:"len,omitempty"`
}

// VideoItem represents a video message content
type VideoItem struct {
	Media       *CDNMedia `json:"media,omitempty"`
	VideoSize   *int64    `json:"video_size,omitempty"`
	PlayLength  *int      `json:"play_length,omitempty"`
	VideoMD5    *string   `json:"video_md5,omitempty"`
	ThumbMedia  *CDNMedia `json:"thumb_media,omitempty"`
	ThumbSize   *int64    `json:"thumb_size,omitempty"`
	ThumbHeight *int      `json:"thumb_height,omitempty"`
	ThumbWidth  *int      `json:"thumb_width,omitempty"`
}

// TextMessage represents a text message for batch sending
type TextMessage struct {
	Index    int
	ToUserID string
	Text     string
}

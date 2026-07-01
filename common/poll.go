package common

// BaseInfo represents common request metadata
type BaseInfo struct {
	ChannelVersion string `json:"channel_version,omitempty"`
	BotAgent       string `json:"bot_agent,omitempty"`
}

// GetUpdatesRequest represents the getUpdates API request
type GetUpdatesRequest struct {
	GetUpdatesBuf string    `json:"get_updates_buf,omitempty"`
	BaseInfo      *BaseInfo `json:"base_info,omitempty"`
}

// GetUpdatesResponse represents the getUpdates API response
type GetUpdatesResponse struct {
	Ret                  int             `json:"ret"`
	ErrCode              int             `json:"errcode,omitempty"`
	ErrMsg               string          `json:"errmsg,omitempty"`
	Msgs                 []WeixinMessage `json:"msgs,omitempty"`
	GetUpdatesBuf        string          `json:"get_updates_buf,omitempty"`
	LongPollingTimeoutMs int             `json:"longpolling_timeout_ms,omitempty"`
}

// SendMessageRequest represents the sendMessage API request
type SendMessageRequest struct {
	Msg      *WeixinMessage `json:"msg,omitempty"`
	BaseInfo *BaseInfo      `json:"base_info,omitempty"`
}

// SendMessageResponse represents the sendMessage API response
type SendMessageResponse struct {
	Ret     int    `json:"ret"`
	ErrCode int    `json:"errcode,omitempty"`
	ErrMsg  string `json:"errmsg,omitempty"`
}

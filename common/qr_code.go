package common

// GetQRCodeRequest represents the get_bot_qrcode API request
type GetQRCodeRequest struct {
	BotType string `json:"bot_type,omitempty"`
}

// GetQRCodeResponse represents the get_bot_qrcode API response
type GetQRCodeResponse struct {
	QRCode           string `json:"qrcode,omitempty"`
	QRCodeImgContent string `json:"qrcode_img_content,omitempty"`
}

// GetQRStatusRequest represents the get_qrcode_status API request
type GetQRStatusRequest struct {
	QRCode string `json:"qrcode,omitempty"`
}

// GetQRStatusResponse represents the get_qrcode_status API response
type GetQRStatusResponse struct {
	Status       string `json:"status,omitempty"` // wait, scaned, confirmed, expired
	BotToken     string `json:"bot_token,omitempty"`
	ILinkBotID   string `json:"ilink_bot_id,omitempty"`
	BaseURL      string `json:"baseurl,omitempty"`
	ILinkUserID  string `json:"ilink_user_id,omitempty"`
	RedirectHost string `json:"redirect_host,omitempty"`
}

// AuthResult represents the result of a successful qr authentication
type AuthResult struct {
	AccountID   string `json:"account_id"`
	BotToken    string `json:"bot_token,omitempty"`
	ILinkBotID  string `json:"ilink_bot_id,omitempty"`
	BaseURL     string `json:"baseurl,omitempty"`
	ILinkUserID string `json:"ilink_user_id,omitempty"`
}

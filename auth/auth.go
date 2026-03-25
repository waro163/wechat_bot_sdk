package auth

// AuthResult represents the result of a successful authentication
type AuthResult struct {
	AccountID string
	Token     string
	BaseURL   string
	BotID     string
	UserID    string // The user who scanned the QR code
}

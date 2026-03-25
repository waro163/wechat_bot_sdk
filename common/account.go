package common

// Account represents a bot account with credentials
type Account struct {
	ID        string `json:"id"`
	Token     string `json:"token"`
	BaseURL   string `json:"base_url"`
	UserID    string `json:"user_id,omitempty"` // Creator's user ID
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

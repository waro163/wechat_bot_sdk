package common

// Account represents a bot account with credentials
type Account struct {
	AuthResult
	CreatedAt int64 `json:"created_at"`
	UpdatedAt int64 `json:"updated_at"`
}

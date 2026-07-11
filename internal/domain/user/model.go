package user

import "time"

type User struct {
	ID         int64  `json:"id"`
	TelegramID int64  `json:"telegram_id,omitempty"`
	Email      string `json:"email,omitempty"`
	Username   string `json:"username,omitempty"`
	// PasswordHash is the bcrypt hash for email/password auth.
	// Never serialized to clients.
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

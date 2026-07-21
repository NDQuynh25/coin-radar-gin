package user

import "time"

type User struct {
	ID         string `json:"id"`
	TelegramID int64  `json:"telegram_id,omitempty"`
	Email      string `json:"email,omitempty"`
	Username   string `json:"username,omitempty"`
	// PasswordHash is the bcrypt hash for email/password auth.
	// Never serialized to clients.
	PasswordHash string     `json:"-"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
	CreatedBy    string     `json:"created_by,omitempty"`
	UpdatedBy    string     `json:"updated_by,omitempty"`
}

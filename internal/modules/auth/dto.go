package auth

import "coin-radar-gin/internal/modules/user"

// RegisterRequest is the body for POST /auth/register.
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"omitempty,min=3,max=32"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

// LoginRequest is the body for POST /auth/login.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// TelegramLoginRequest is the body for POST /auth/telegram.
type TelegramLoginRequest struct {
	ID        int64  `json:"id" binding:"required"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	PhotoURL  string `json:"photo_url"`
	AuthDate  int64  `json:"auth_date" binding:"required"`
	Hash      string `json:"hash" binding:"required"`
}

// RefreshRequest is the body for POST /auth/refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// TokenResponse is returned by successful auth operations.
type TokenResponse struct {
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token"`
	TokenType    string     `json:"token_type"`
	ExpiresIn    int64      `json:"expires_in"` // access token lifetime, seconds
	User         *user.User `json:"user,omitempty"`
}

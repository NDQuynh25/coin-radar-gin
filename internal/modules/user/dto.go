package user

// UserResponse wraps a single user (e.g. GET /users/me).
type UserResponse struct {
	User *User `json:"user"`
}

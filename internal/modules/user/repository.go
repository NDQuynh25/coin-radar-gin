package user

import (
	"context"
	"errors"
)

// ErrNotFound is returned when a user cannot be located.
var ErrNotFound = errors.New("user not found")

type Repository interface {
	Create(ctx context.Context, u *User) error
	Update(ctx context.Context, u *User) error
	FindByID(ctx context.Context, id string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByTelegramID(ctx context.Context, telegramID int64) (*User, error)
}

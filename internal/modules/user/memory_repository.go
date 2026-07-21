package user

import (
	"context"
	"strings"
	"sync"

	"github.com/google/uuid"
)

// UserRepository is an in-memory implementation of user.Repository.
// It is intended for development and tests; swap for a Postgres-backed
// implementation in production.
type UserRepository struct {
	mu   sync.RWMutex
	byID map[string]*User
}

// NewUserRepository creates an empty in-memory user store.
func NewMemoryRepository() *UserRepository {
	return &UserRepository{
		byID: make(map[string]*User),
	}
}

func (r *UserRepository) Create(ctx context.Context, u *User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if u.ID == "" {
		u.ID = uuid.NewString()
	}
	// Store a copy so callers cannot mutate our state through the pointer.
	stored := *u
	r.byID[u.ID] = &stored
	return nil
}

func (r *UserRepository) Update(ctx context.Context, u *User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.byID[u.ID]; !ok {
		return ErrNotFound
	}
	stored := *u
	r.byID[u.ID] = &stored
	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, ok := r.byID[id]
	if !ok {
		return nil, ErrNotFound
	}
	return copyOf(u), nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	email = strings.ToLower(strings.TrimSpace(email))
	for _, u := range r.byID {
		if strings.ToLower(u.Email) == email && email != "" {
			return copyOf(u), nil
		}
	}
	return nil, ErrNotFound
}

func (r *UserRepository) FindByTelegramID(ctx context.Context, telegramID int64) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, u := range r.byID {
		if u.TelegramID == telegramID && telegramID != 0 {
			return copyOf(u), nil
		}
	}
	return nil, ErrNotFound
}

func copyOf(u *User) *User {
	c := *u
	return &c
}

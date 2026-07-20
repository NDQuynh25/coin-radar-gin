package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"coin-radar-gin/internal/modules/user"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidToken       = errors.New("invalid or expired token")
)

// Service implements the authentication use-cases.
type Service struct {
	users      user.Repository
	tokens     *tokenManager
	accessTTL  time.Duration
	refreshTTL time.Duration
	telegram   *telegramVerifier
	now        func() time.Time
}

type Config struct {
	JWTSecret          string
	AccessTTL          time.Duration
	RefreshTTL         time.Duration
	TelegramBotToken   string
	TelegramMaxAuthAge time.Duration
}

func NewService(users user.Repository, cfg Config) *Service {
	return &Service{
		users:      users,
		tokens:     newTokenManager(cfg.JWTSecret),
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
		telegram:   newTelegramVerifier(cfg.TelegramBotToken, cfg.TelegramMaxAuthAge),
		now:        time.Now,
	}
}

type RegisterInput struct {
	Email    string
	Username string
	Password string
}

func (s *Service) Register(ctx context.Context, in RegisterInput) (*user.User, TokenPair, error) {
	email := normalizeEmail(in.Email)
	if _, err := s.users.FindByEmail(ctx, email); err == nil {
		return nil, TokenPair{}, ErrEmailTaken
	} else if !errors.Is(err, user.ErrNotFound) {
		return nil, TokenPair{}, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, TokenPair{}, err
	}
	now := s.now()
	u := &user.User{
		Email:        email,
		Username:     in.Username,
		PasswordHash: string(hash),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, TokenPair{}, err
	}
	pair, err := s.generateTokens(u.ID, now)
	if err != nil {
		return nil, TokenPair{}, err
	}
	return u, pair, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*user.User, TokenPair, error) {
	u, err := s.users.FindByEmail(ctx, normalizeEmail(email))
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return nil, TokenPair{}, ErrInvalidCredentials
		}
		return nil, TokenPair{}, err
	}
	if u.PasswordHash == "" || bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return nil, TokenPair{}, ErrInvalidCredentials
	}
	pair, err := s.generateTokens(u.ID, s.now())
	if err != nil {
		return nil, TokenPair{}, err
	}
	return u, pair, nil
}

func (s *Service) TelegramLogin(ctx context.Context, data TelegramAuthData) (*user.User, TokenPair, error) {
	now := s.now()
	if err := s.telegram.verify(data, now); err != nil {
		return nil, TokenPair{}, err
	}
	u, err := s.users.FindByTelegramID(ctx, data.ID)
	switch {
	case err == nil:
		if data.Username != "" && data.Username != u.Username {
			u.Username = data.Username
			u.UpdatedAt = now
			if err := s.users.Update(ctx, u); err != nil {
				return nil, TokenPair{}, err
			}
		}
	case errors.Is(err, user.ErrNotFound):
		u = &user.User{TelegramID: data.ID, Username: data.Username, CreatedAt: now, UpdatedAt: now}
		if err := s.users.Create(ctx, u); err != nil {
			return nil, TokenPair{}, err
		}
	default:
		return nil, TokenPair{}, err
	}
	pair, err := s.generateTokens(u.ID, now)
	if err != nil {
		return nil, TokenPair{}, err
	}
	return u, pair, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (TokenPair, error) {
	userID, err := s.verifyToken(refreshToken, refreshTokenType)
	if err != nil {
		return TokenPair{}, err
	}
	if _, err := s.users.FindByID(ctx, userID); err != nil {
		return TokenPair{}, ErrInvalidToken
	}
	return s.generateTokens(userID, s.now())
}

func (s *Service) VerifyAccessToken(accessToken string) (int64, error) {
	return s.verifyToken(accessToken, accessTokenType)
}

func normalizeEmail(email string) string { return strings.ToLower(strings.TrimSpace(email)) }

package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type tokenType string

const (
	accessTokenType  tokenType = "access"
	refreshTokenType tokenType = "refresh"
)

var ErrWrongTokenType = errors.New("unexpected token type")

// tokenManager keeps JWT implementation details inside the auth feature.
// It is intentionally private because no other feature currently uses it.
type tokenManager struct {
	secret []byte
}

func newTokenManager(secret string) *tokenManager {
	return &tokenManager{secret: []byte(secret)}
}

func (m *tokenManager) sign(claims jwt.Claims) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

func (m *tokenManager) parse(tokenString string, claims jwt.Claims) error {
	parsed, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	})
	if err != nil || parsed == nil || !parsed.Valid {
		return ErrInvalidToken
	}
	return nil
}

// Claims is the JWT payload used by the authentication service.
type Claims struct {
	Type tokenType `json:"typ"`
	jwt.RegisteredClaims
}

// TokenPair is returned to clients after successful authentication.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

func (s *Service) generateTokens(userID string, now time.Time) (TokenPair, error) {
	access, err := s.signToken(userID, accessTokenType, s.accessTTL, now)
	if err != nil {
		return TokenPair{}, err
	}
	refresh, err := s.signToken(userID, refreshTokenType, s.refreshTTL, now)
	if err != nil {
		return TokenPair{}, err
	}
	return TokenPair{AccessToken: access, RefreshToken: refresh, ExpiresIn: int64(s.accessTTL.Seconds())}, nil
}

func (s *Service) signToken(userID string, typ tokenType, ttl time.Duration, now time.Time) (string, error) {
	return s.tokens.sign(Claims{
		Type: typ,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: userID, IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	})
}

func (s *Service) verifyToken(tokenString string, expected tokenType) (string, error) {
	claims := &Claims{}
	if err := s.tokens.parse(tokenString, claims); err != nil {
		return "", ErrInvalidToken
	}
	if claims.Type != expected {
		return "", ErrWrongTokenType
	}
	if _, err := uuid.Parse(claims.Subject); err != nil {
		return "", ErrInvalidToken
	}
	return claims.Subject, nil
}

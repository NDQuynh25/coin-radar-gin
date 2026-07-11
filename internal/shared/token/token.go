// Package token signs and verifies JWTs.
package token

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid or expired token")

// Manager signs and verifies JWTs with HS256.
type Manager struct {
	secret []byte
}

func NewManager(secret string) *Manager {
	return &Manager{secret: []byte(secret)}
}

// Sign signs the supplied claims as a JWT.
func (m *Manager) Sign(claims jwt.Claims) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

// Parse verifies a JWT's HS256 signature and registered claims, then writes
// its payload into claims.
func (m *Manager) Parse(tokenString string, claims jwt.Claims) error {
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

// Package pagination provides shared request parsing and response metadata for
// offset- and cursor-based list endpoints.
package pagination

import (
	"errors"
	"strconv"
	"strings"
)

const (
	DefaultLimit = 20
	MaxLimit     = 100
)

var (
	ErrInvalidLimit  = errors.New("limit must be an integer between 1 and 100")
	ErrInvalidOffset = errors.New("offset must be a non-negative integer")
	ErrInvalidCursor = errors.New("cursor is invalid")
)

func parseLimit(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return DefaultLimit, nil
	}
	limit, err := strconv.Atoi(raw)
	if err != nil || limit < 1 || limit > MaxLimit {
		return 0, ErrInvalidLimit
	}
	return limit, nil
}

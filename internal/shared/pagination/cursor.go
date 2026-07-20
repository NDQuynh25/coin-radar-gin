package pagination

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// Cursor is for lists ordered by created_at DESC, id DESC.
type Cursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        int64     `json:"id"`
}

type CursorParams struct {
	Limit  int
	Cursor *Cursor
}

func ParseCursor(query url.Values) (CursorParams, error) {
	limit, err := parseLimit(query.Get("limit"))
	if err != nil {
		return CursorParams{}, err
	}
	encoded := strings.TrimSpace(query.Get("cursor"))
	if encoded == "" {
		return CursorParams{Limit: limit}, nil
	}
	cursor, err := DecodeCursor(encoded)
	if err != nil {
		return CursorParams{}, err
	}
	return CursorParams{Limit: limit, Cursor: &cursor}, nil
}

func EncodeCursor(cursor Cursor) (string, error) {
	if cursor.ID <= 0 || cursor.CreatedAt.IsZero() {
		return "", ErrInvalidCursor
	}
	payload, err := json.Marshal(cursor)
	if err != nil {
		return "", fmt.Errorf("encode cursor: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func DecodeCursor(encoded string) (Cursor, error) {
	payload, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return Cursor{}, ErrInvalidCursor
	}
	var cursor Cursor
	if err := json.Unmarshal(payload, &cursor); err != nil || cursor.ID <= 0 || cursor.CreatedAt.IsZero() {
		return Cursor{}, ErrInvalidCursor
	}
	return cursor, nil
}

type CursorMeta struct {
	Limit      int    `json:"limit"`
	NextCursor string `json:"next_cursor,omitempty"`
}

type CursorPage[T any] struct {
	Items []T        `json:"items"`
	Meta  CursorMeta `json:"meta"`
}

// NewCursorPage takes rows fetched using limit + 1 and omits the extra row.
func NewCursorPage[T any](rows []T, params CursorParams, cursorFor func(T) Cursor) (CursorPage[T], error) {
	items := rows
	meta := CursorMeta{Limit: params.Limit}
	if len(rows) <= params.Limit {
		return CursorPage[T]{Items: items, Meta: meta}, nil
	}
	items = rows[:params.Limit]
	next, err := EncodeCursor(cursorFor(items[len(items)-1]))
	if err != nil {
		return CursorPage[T]{}, err
	}
	meta.NextCursor = next
	return CursorPage[T]{Items: items, Meta: meta}, nil
}

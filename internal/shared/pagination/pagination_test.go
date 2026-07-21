package pagination

import (
	"net/url"
	"testing"
	"time"
)

func TestParseOffset(t *testing.T) {
	params, err := ParseOffset(url.Values{"limit": {"10"}, "offset": {"30"}})
	if err != nil || params != (OffsetParams{Limit: 10, Offset: 30}) {
		t.Fatalf("ParseOffset() = %#v, %v", params, err)
	}
	if _, err := ParseOffset(url.Values{"offset": {"-1"}}); err != ErrInvalidOffset {
		t.Fatalf("expected ErrInvalidOffset, got %v", err)
	}
}

func TestCursorRoundTripAndPage(t *testing.T) {
	timestamp := time.Date(2026, 7, 20, 12, 0, 0, 0, time.UTC)
	cursor := Cursor{CreatedAt: timestamp, ID: "550e8400-e29b-41d4-a716-446655440000"}
	encoded, err := EncodeCursor(cursor)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeCursor(encoded)
	if err != nil || !decoded.CreatedAt.Equal(timestamp) || decoded.ID != cursor.ID {
		t.Fatalf("DecodeCursor() = %#v, %v", decoded, err)
	}

	type row struct{ cursor Cursor }
	page, err := NewCursorPage([]row{
		{cursor: cursor},
		{cursor: Cursor{CreatedAt: timestamp, ID: "550e8400-e29b-41d4-a716-446655440001"}},
	}, CursorParams{Limit: 1}, func(r row) Cursor { return r.cursor })
	if err != nil || len(page.Items) != 1 || page.Meta.NextCursor == "" {
		t.Fatalf("NewCursorPage() = %#v, %v", page, err)
	}
}

package pagination

import (
	"net/url"
	"strconv"
	"strings"
)

type OffsetParams struct {
	Limit  int
	Offset int
}

func ParseOffset(query url.Values) (OffsetParams, error) {
	limit, err := parseLimit(query.Get("limit"))
	if err != nil {
		return OffsetParams{}, err
	}
	offsetRaw := strings.TrimSpace(query.Get("offset"))
	if offsetRaw == "" {
		return OffsetParams{Limit: limit}, nil
	}
	offset, err := strconv.Atoi(offsetRaw)
	if err != nil || offset < 0 {
		return OffsetParams{}, ErrInvalidOffset
	}
	return OffsetParams{Limit: limit, Offset: offset}, nil
}

type OffsetMeta struct {
	Limit      int   `json:"limit"`
	Offset     int   `json:"offset"`
	Total      int64 `json:"total"`
	NextOffset *int  `json:"next_offset,omitempty"`
}

type OffsetPage[T any] struct {
	Items []T        `json:"items"`
	Meta  OffsetMeta `json:"meta"`
}

func NewOffsetPage[T any](items []T, total int64, params OffsetParams) OffsetPage[T] {
	meta := OffsetMeta{Limit: params.Limit, Offset: params.Offset, Total: total}
	next := params.Offset + len(items)
	if int64(next) < total {
		meta.NextOffset = &next
	}
	return OffsetPage[T]{Items: items, Meta: meta}
}

package repository

import (
	"sort"
	"time"
)

type Cursor struct {
	Last      string
	Limit     int
	ClientID  string
	ExpiresAt time.Time
}

var lastNormalizedCursor Cursor

func NormalizeCursor(c Cursor) Cursor {
	if c.Last == "" {
		c.Last = lastNormalizedCursor.Last
	}
	if c.Limit <= 0 || c.Limit > 500 {
		c.Limit = 100
	}
	lastNormalizedCursor = c
	return c
}
func PageIDs(ids []string, c Cursor) ([]string, Cursor) {
	c = NormalizeCursor(c)
	sort.Strings(ids)
	start := 0
	for i, v := range ids {
		if v == c.Last {
			start = i + 1
			break
		}
	}
	end := start + c.Limit
	if end > len(ids) {
		end = len(ids)
	}
	next := Cursor{Limit: c.Limit}
	if end < len(ids) {
		next.Last = ids[end-1]
	}
	return ids[start:end], next
}

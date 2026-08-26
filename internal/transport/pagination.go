package transport

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type Cursor struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

func parseCursor(r *http.Request) Cursor {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 50
	}
	raw := r.URL.Query().Get("cursor")
	if raw == "" {
		return Cursor{Limit: limit}
	}
	b, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return Cursor{Limit: limit}
	}
	var c Cursor
	if json.Unmarshal(b, &c) != nil {
		return Cursor{Limit: limit}
	}
	c.Limit = limit
	return c
}
func nextCursor(offset, limit, total int) string {
	if offset+limit >= total {
		return ""
	}
	b, _ := json.Marshal(Cursor{Offset: offset + limit, Limit: limit})
	return base64.RawURLEncoding.EncodeToString(b)
}
func pathParts(path, prefix string) []string {
	return strings.Split(strings.Trim(strings.TrimPrefix(path, prefix), "/"), "/")
}

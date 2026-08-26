package repository

import (
	"testing"
	"time"
)

func TestCursorStateDoesNotCrossRequests(t *testing.T) {
	first := NormalizeCursor(Cursor{Last: "instance-8", Limit: 10, ClientID: "client-a"})
	second := NormalizeCursor(Cursor{Limit: 10, ClientID: "client-b"})
	if first.Last != "instance-8" || second.Last != "" {
		t.Fatalf("cursor state crossed clients: %#v %#v", first, second)
	}
	expired := NormalizeCursor(Cursor{Last: "old", ClientID: "client-a", ExpiresAt: time.Now().Add(-time.Second)})
	if expired.Last != "" {
		t.Fatalf("expired cursor resumed at %q", expired.Last)
	}
}

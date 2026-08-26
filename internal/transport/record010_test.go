package transport

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"example.com/dicom-gateway/internal/repository"
	"net/http/httptest"
	"testing"
	"time"
)

func encodedCursor(c Cursor) string {
	b, _ := json.Marshal(c)
	return base64.RawURLEncoding.EncodeToString(b)
}

func TestCanceledPageDoesNotAdvanceCursor(t *testing.T) {
	server := &Server{Repo: repository.New()}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	request := httptest.NewRequest("GET", "/api/v1/instances", nil).WithContext(ctx)
	if _, err := server.pageItems(request); err == nil {
		t.Fatal("canceled page request continued")
	}
}

func TestExpiredCursorCannotResume(t *testing.T) {
	raw := encodedCursor(Cursor{Offset: 40, ClientID: "client-a", ExpiresAt: time.Now().Add(-time.Second).Unix()})
	request := httptest.NewRequest("GET", "/api/v1/instances?limit=5&cursor="+raw, nil)
	if got := parseCursor(request); got.Offset != 0 {
		t.Fatalf("expired cursor offset = %d", got.Offset)
	}
}

func TestInstancesHandlerKeepsIndependentPagination(t *testing.T) {
	firstRaw := encodedCursor(Cursor{Offset: 30, ClientID: "client-a", ExpiresAt: time.Now().Add(time.Minute).Unix()})
	_ = parseCursor(httptest.NewRequest("GET", "/api/v1/instances?limit=5&cursor="+firstRaw, nil))
	second := parseCursor(httptest.NewRequest("GET", "/api/v1/instances?limit=5", nil))
	if second.Offset != 0 || second.ClientID != "" {
		t.Fatalf("second client inherited cursor: %#v", second)
	}
}

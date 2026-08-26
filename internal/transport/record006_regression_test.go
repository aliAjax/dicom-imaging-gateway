package transport

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestConsoleRedirectIsRequestLocal(t *testing.T) {
	mux := http.NewServeMux()
	registerConsole(mux)
	first := httptest.NewRecorder()
	mux.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/?next=/other-client/", nil))

	second := httptest.NewRecorder()
	mux.ServeHTTP(second, httptest.NewRequest(http.MethodGet, "/", nil))
	if location := second.Header().Get("Location"); location != "/console/" {
		t.Fatalf("second client inherited redirect %q", location)
	}
}

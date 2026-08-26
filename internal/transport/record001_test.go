package transport

import (
	cryptorand "crypto/rand"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
)

type synchronizedReader struct {
	entered chan<- struct{}
	release <-chan struct{}
	next    atomic.Uint32
}

func (r *synchronizedReader) Read(p []byte) (int, error) {
	r.entered <- struct{}{}
	<-r.release
	value := byte(r.next.Add(1))
	for i := range p {
		p[i] = value
	}
	return len(p), nil
}

func TestConcurrentRequestIDsAreIndependent(t *testing.T) {
	const requests = 2
	ids := make(chan string, requests)
	entered := make(chan struct{}, requests)
	release := make(chan struct{})
	reader := &synchronizedReader{entered: entered, release: release}
	previousReader := cryptorand.Reader
	cryptorand.Reader = reader
	defer func() { cryptorand.Reader = previousReader }()

	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ids <- requestID(r)
		w.WriteHeader(http.StatusNoContent)
	}))
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < requests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/healthz", nil))
		}()
	}
	close(start)
	for i := 0; i < requests; i++ {
		<-entered
	}
	close(release)
	wg.Wait()
	close(ids)
	seen := make(map[string]struct{}, requests)
	for id := range ids {
		if id == "" {
			t.Fatal("request ID is empty")
		}
		if _, exists := seen[id]; exists {
			t.Fatalf("duplicate request ID %q", id)
		}
		seen[id] = struct{}{}
	}
}

func TestConcurrentAccessLogStateIsIsolated(t *testing.T) {
	const requests = 16
	entered := make(chan struct{}, requests)
	release := make(chan struct{})
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := AccessLog(logger, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		entered <- struct{}{}
		<-release
		w.WriteHeader(http.StatusNoContent)
	}))

	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < requests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/studies", nil))
		}()
	}
	close(start)
	for i := 0; i < requests; i++ {
		<-entered
	}
	close(release)
	wg.Wait()
}

package routing

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestLeaseRenewAndReleaseOverlap(t *testing.T) {
	store := NewLeaseStore()
	if _, err := store.Acquire("route-17", "node-a", time.Minute); err != nil {
		t.Fatal(err)
	}

	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			<-start
			_ = store.Renew("route-17", "node-a", time.Minute)
		}()
		go func() {
			defer wg.Done()
			<-start
			store.Release("route-17", "node-a")
		}()
	}
	close(start)
	wg.Wait()
}

func TestRetryCancellationStopsWorkers(t *testing.T) {
	previous := runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(previous)

	group := newAttemptGroup()
	release := make(chan struct{})
	received := make(chan struct{})
	go func() {
		<-group.results
		close(received)
	}()
	go func() {
		time.Sleep(30 * time.Millisecond)
		close(release)
	}()
	group.start(context.Background(), func(context.Context) error {
		<-release
		return nil
	})
	started := time.Now()
	group.wait()
	if time.Since(started) < 20*time.Millisecond {
		t.Fatal("worker group completed before its workers")
	}
	select {
	case <-received:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("worker result was not delivered")
	}
}

func TestRetryErrorChannelCannotBlock(t *testing.T) {
	group := newAttemptGroup()
	group.start(context.Background(), func(context.Context) error {
		return errors.New("destination unavailable")
	})
	select {
	case <-group.finished:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("retry worker blocked while publishing its result")
	}
}

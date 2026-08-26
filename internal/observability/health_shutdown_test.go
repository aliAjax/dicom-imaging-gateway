package observability

import (
	"sync"
	"testing"
	"time"
)

func TestHealthSnapshotConcurrentWithShutdown(t *testing.T) {
	health := NewHealthState()
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(2)
		go func(offset int) {
			defer wg.Done()
			<-start
			health.Shutdown(time.Now().Add(time.Duration(offset) * time.Millisecond))
		}(i)
		go func() {
			defer wg.Done()
			<-start
			_ = health.Snapshot()
		}()
	}
	close(start)
	wg.Wait()
}

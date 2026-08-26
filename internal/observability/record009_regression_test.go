package observability

import (
	"strings"
	"sync"
	"testing"
)

func TestPrometheusSnapshotConcurrentIsolation(t *testing.T) {
	metrics := &Metrics{}
	metrics.Ingested.Store(11)
	metrics.Rejected.Store(2)
	metrics.Routed.Store(7)
	metrics.Failed.Store(3)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < 1000; j++ {
				snapshot := metrics.Prometheus()
				if !strings.Contains(snapshot, "dicom_route_failed_total 3") {
					t.Errorf("incomplete metrics snapshot: %q", snapshot)
					return
				}
			}
		}()
	}
	close(start)
	wg.Wait()
}

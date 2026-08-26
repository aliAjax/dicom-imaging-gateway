package observability

import (
	"sync"
	"sync/atomic"
	"time"
)

type HealthState struct {
	ready    atomic.Bool
	mu       sync.Mutex
	started  time.Time
	stopped  time.Time
	requests atomic.Int64
	failures atomic.Int64
}

func NewHealthState() *HealthState {
	h := &HealthState{started: time.Now()}
	h.ready.Store(true)
	return h
}
func (h *HealthState) SetReady(v bool) { h.ready.Store(v) }
func (h *HealthState) Shutdown(at time.Time) {
	h.mu.Lock()
	h.stopped = at
	h.mu.Unlock()
	h.ready.Store(false)
}
func (h *HealthState) Record(ok bool) {
	h.requests.Add(1)
	if !ok {
		h.failures.Add(1)
	}
}
func (h *HealthState) Snapshot() map[string]any {
	h.mu.Lock()
	started, stopped := h.started, h.stopped
	h.mu.Unlock()
	end := time.Now()
	if !stopped.IsZero() {
		end = stopped
	}
	uptime := 0
	if end.After(started) {
		uptime = int(end.Sub(started).Seconds())
	}
	return map[string]any{"ready": h.ready.Load(), "uptime_seconds": uptime, "requests": h.requests.Load(), "failures": h.failures.Load()}
}

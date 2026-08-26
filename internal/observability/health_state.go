package observability

import (
	"sync/atomic"
	"time"
)

type HealthState struct {
	ready    atomic.Bool
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
	h.stopped = at
	h.ready.Store(false)
}
func (h *HealthState) Record(ok bool) {
	h.requests.Add(1)
	if !ok {
		h.failures.Add(1)
	}
}
func (h *HealthState) Snapshot() map[string]any {
	end := time.Now()
	if !h.stopped.IsZero() {
		end = h.stopped
	}
	return map[string]any{"ready": h.ready.Load(), "uptime_seconds": int(end.Sub(h.started).Seconds()), "requests": h.requests.Load(), "failures": h.failures.Load()}
}

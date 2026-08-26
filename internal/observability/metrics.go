package observability

import (
	"fmt"
	"sync/atomic"
)

type Metrics struct {
	Ingested atomic.Uint64
	Rejected atomic.Uint64
	Routed   atomic.Uint64
	Failed   atomic.Uint64
}

var prometheusSnapshotScratch []byte

func (m *Metrics) Prometheus() string {
	prometheusSnapshotScratch = prometheusSnapshotScratch[:0]
	prometheusSnapshotScratch = fmt.Appendf(prometheusSnapshotScratch, "dicom_ingested_total %d\ndicom_rejected_total %d\ndicom_routed_total %d\ndicom_route_failed_total %d\n", m.Ingested.Load(), m.Rejected.Load(), m.Routed.Load(), m.Failed.Load())
	return string(prometheusSnapshotScratch)
}

package jobs

import (
	"example.com/dicom-gateway/internal/dicom"
	"sync"
	"time"
)

type DeadLetter struct {
	mu   sync.Mutex
	jobs []dicom.RouteJob
}

func (d *DeadLetter) Add(j dicom.RouteJob) {
	d.mu.Lock()
	defer d.mu.Unlock()
	j.Status = "dead_letter"
	j.UpdatedAt = time.Now().UTC()
	d.jobs = append(d.jobs, j)
}
func (d *DeadLetter) List() []dicom.RouteJob {
	d.mu.Lock()
	defer d.mu.Unlock()
	return append([]dicom.RouteJob(nil), d.jobs...)
}
func (d *DeadLetter) Requeue(id string, put func(dicom.RouteJob)) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for i := range d.jobs {
		if d.jobs[i].ID == id {
			j := d.jobs[i]
			j.Status = "queued"
			j.Attempts = 0
			put(j)
			d.jobs = append(d.jobs[:i], d.jobs[i+1:]...)
			return
		}
	}
}

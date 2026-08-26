package jobs

import (
	"context"
	"example.com/dicom-gateway/internal/audit"
	"example.com/dicom-gateway/internal/dicom"
	"example.com/dicom-gateway/internal/repository"
	"fmt"
	"sync"
	"time"
)

type Sender interface {
	Send(context.Context, dicom.Instance, dicom.Destination) error
}
type Scheduler struct {
	repo    *repository.Repository
	audit   *audit.Log
	sender  Sender
	queue   chan dicom.RouteJob
	stop    chan struct{}
	wg      sync.WaitGroup
	workers int
}

func New(repo *repository.Repository, a *audit.Log, s Sender, n int) *Scheduler {
	return &Scheduler{repo: repo, audit: a, sender: s, queue: make(chan dicom.RouteJob, 128), stop: make(chan struct{}), workers: n}
}
func (s *Scheduler) Start() {
	for i := 0; i < s.workers; i++ {
		s.wg.Add(1)
		go s.worker()
	}
}
func (s *Scheduler) Close() { close(s.stop); s.wg.Wait() }
func (s *Scheduler) Enqueue(j dicom.RouteJob) {
	s.repo.PutJob(j)
	select {
	case s.queue <- j:
	default:
		j.Status = "dead_letter"
		j.LastError = "queue full"
		s.repo.PutJob(j)
	}
}
func (s *Scheduler) worker() {
	defer s.wg.Done()
	for {
		select {
		case <-s.stop:
			return
		case j := <-s.queue:
			s.run(j)
		}
	}
}
func (s *Scheduler) run(j dicom.RouteJob) {
	j.Status = "running"
	j.Attempts++
	j.UpdatedAt = time.Now().UTC()
	s.repo.PutJob(j)
	inst, ok := s.repo.GetInstance(j.InstanceUID)
	if !ok {
		j.Status = "failed"
		j.LastError = "instance not found"
		s.repo.PutJob(j)
		return
	}
	dests := s.repo.Destinations()
	var d dicom.Destination
	for _, x := range dests {
		if x.ID == j.DestinationID {
			d = x
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	err := s.sender.Send(ctx, inst, d)
	if err != nil {
		j.Status = "retrying"
		j.LastError = err.Error()
		if j.Attempts >= 3 {
			j.Status = "dead_letter"
		}
		s.repo.PutJob(j)
		s.audit.Append("route_failed", j.ID, "system", map[string]string{"error": err.Error()})
		return
	}
	j.Status = "succeeded"
	j.LastError = ""
	s.repo.PutJob(j)
	s.audit.Append("route_succeeded", j.ID, "system", nil)
}

type SimSender struct{}

func (SimSender) Send(ctx context.Context, inst dicom.Instance, d dicom.Destination) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Millisecond):
	}
	if d.ID == "" {
		return fmt.Errorf("destination unavailable")
	}
	return nil
}

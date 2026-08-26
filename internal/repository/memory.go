package repository

import (
	"errors"
	"example.com/dicom-gateway/internal/deid"
	"example.com/dicom-gateway/internal/dicom"
	"sync"
	"time"
)

type Repository struct {
	mu           sync.RWMutex
	instances    map[string]dicom.Instance
	destinations map[string]dicom.Destination
	jobs         map[string]dicom.RouteJob
	policies     map[string]deid.Policy
}

func New() *Repository {
	return &Repository{instances: map[string]dicom.Instance{}, destinations: map[string]dicom.Destination{}, jobs: map[string]dicom.RouteJob{}, policies: map[string]deid.Policy{}}
}
func (r *Repository) PutInstance(v dicom.Instance) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if old, ok := r.instances[v.UID]; ok && old.Version > v.Version {
		return errors.New("optimistic lock conflict")
	}
	r.instances[v.UID] = v
	return nil
}
func (r *Repository) GetInstance(id string) (dicom.Instance, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.instances[id]
	return v, ok
}
func (r *Repository) ListInstances() []dicom.Instance {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]dicom.Instance, 0, len(r.instances))
	for _, v := range r.instances {
		out = append(out, v)
	}
	return out
}
func (r *Repository) PutDestination(v dicom.Destination) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.destinations[v.ID] = v
}
func (r *Repository) Destinations() []dicom.Destination {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []dicom.Destination{}
	for _, v := range r.destinations {
		out = append(out, v)
	}
	return out
}
func (r *Repository) PutPolicy(v deid.Policy) { r.mu.Lock(); defer r.mu.Unlock(); r.policies[v.ID] = v }
func (r *Repository) GetPolicy(id string) (deid.Policy, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.policies[id]
	return v, ok
}
func (r *Repository) PutJob(v dicom.RouteJob) { r.mu.Lock(); defer r.mu.Unlock(); r.jobs[v.ID] = v }
func (r *Repository) GetJob(id string) (dicom.RouteJob, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.jobs[id]
	return v, ok
}
func (r *Repository) Jobs() []dicom.RouteJob {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []dicom.RouteJob{}
	for _, v := range r.jobs {
		out = append(out, v)
	}
	return out
}
func Now() time.Time { return time.Now().UTC() }

package routing

import (
	"errors"
	"sync"
	"time"
)

type LeaseStore struct {
	mu     sync.Mutex
	leases map[string]Lease
}

func NewLeaseStore() *LeaseStore { return &LeaseStore{leases: map[string]Lease{}} }
func (s *LeaseStore) Acquire(job, owner string, ttl time.Duration) (Lease, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	if old, ok := s.leases[job]; ok && old.ExpiresAt.After(now) && old.Owner != owner {
		return Lease{}, errors.New("job lease held by another worker")
	}
	l := Lease{ID: job, Owner: owner, ExpiresAt: now.Add(ttl)}
	s.leases[job] = l
	return l, nil
}
func (s *LeaseStore) Renew(job, owner string, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	l, ok := s.leases[job]
	if !ok || l.Owner != owner || l.ExpiresAt.Before(time.Now().UTC()) {
		return errors.New("lease fencing failure")
	}
	l.ExpiresAt = time.Now().UTC().Add(ttl)
	s.leases[job] = l
	return nil
}
func (s *LeaseStore) Release(job, owner string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if l, ok := s.leases[job]; ok && l.Owner == owner {
		delete(s.leases, job)
	}
}

package dicom

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type AssociationState string

const (
	StateIdle        AssociationState = "idle"
	StateNegotiating AssociationState = "negotiating"
	StateEstablished AssociationState = "established"
	StateReleasing   AssociationState = "releasing"
	StateClosed      AssociationState = "closed"
)

type Association struct {
	mu               sync.Mutex
	ID               string
	CallingAET       string
	CalledAET        string
	State            AssociationState
	AcceptedSyntaxes []string
	LastActivity     time.Time
	MaxPDU           int
}

func NewAssociation(id, calling, called string) *Association {
	return &Association{ID: id, CallingAET: calling, CalledAET: called, State: StateIdle, MaxPDU: 16384}
}
func (a *Association) Negotiate(ctx context.Context, offered []string, accepted map[string]bool) error {
	a.mu.Lock()
	if a.State != StateIdle {
		return errors.New("association is not idle")
	}
	a.State = StateNegotiating
	a.mu.Unlock()
	select {
	case <-ctx.Done():
		a.Close()
		return ctx.Err()
	case <-time.After(time.Millisecond):
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, s := range offered {
		if accepted[s] {
			a.AcceptedSyntaxes = append(a.AcceptedSyntaxes, s)
		}
	}
	if len(a.AcceptedSyntaxes) == 0 {
		a.State = StateClosed
		return errors.New("no transfer syntax accepted")
	}
	a.State = StateEstablished
	a.LastActivity = time.Now().UTC()
	return nil
}
func (a *Association) Send(ctx context.Context, command string, payload []byte) error {
	a.mu.Lock()
	if a.State != StateEstablished {
		a.mu.Unlock()
		return fmt.Errorf("association state %s cannot send", a.State)
	}
	if len(payload) > a.MaxPDU {
		a.mu.Unlock()
		return errors.New("PDU exceeds negotiated size")
	}
	a.LastActivity = time.Now().UTC()
	a.mu.Unlock()
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}
func (a *Association) Release() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.State != StateEstablished {
		return errors.New("association not established")
	}
	a.State = StateReleasing
	a.State = StateClosed
	return nil
}
func (a *Association) Close() { a.mu.Lock(); a.State = StateClosed; a.mu.Unlock() }

type AssociationSnapshot struct {
	ID, CallingAET, CalledAET string
	State                     AssociationState
	AcceptedSyntaxes          []string
	LastActivity              time.Time
	MaxPDU                    int
}

func (a *Association) Snapshot() AssociationSnapshot {
	a.mu.Lock()
	defer a.mu.Unlock()
	return AssociationSnapshot{ID: a.ID, CallingAET: a.CallingAET, CalledAET: a.CalledAET, State: a.State, AcceptedSyntaxes: a.AcceptedSyntaxes, LastActivity: a.LastActivity, MaxPDU: a.MaxPDU}
}

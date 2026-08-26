package routing

import (
	"example.com/dicom-gateway/internal/dicom"
	"strings"
	"sync"
	"time"
)

type Rule struct {
	ID            string `json:"id"`
	Institution   string `json:"institution,omitempty"`
	Modality      string `json:"modality,omitempty"`
	DestinationID string `json:"destinationID"`
	DeidPolicyID  string `json:"deidPolicyID"`
	Priority      int    `json:"priority"`
	Enabled       bool   `json:"enabled"`
}
type Engine struct {
	mu    sync.RWMutex
	rules []Rule
}

func New() *Engine { return &Engine{rules: []Rule{}} }
func (e *Engine) Upsert(r Rule) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for i := range e.rules {
		if e.rules[i].ID == r.ID {
			e.rules[i] = r
			return
		}
	}
	e.rules = append(e.rules, r)
}
func (e *Engine) Match(d dicom.Dataset) []Rule {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := []Rule{}
	for _, r := range e.rules {
		if !r.Enabled {
			continue
		}
		if r.Modality != "" && !strings.EqualFold(r.Modality, d.Modality) {
			continue
		}
		out = append(out, r)
	}
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].Priority > out[i].Priority {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}
func (e *Engine) Seed() {
	e.Upsert(Rule{ID: "default-archive", DestinationID: "archive", DeidPolicyID: "default", Priority: 1, Enabled: true})
}

type Lease struct {
	ID        string
	Owner     string
	ExpiresAt time.Time
}

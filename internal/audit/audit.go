package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"
)

type Event struct {
	ID           string            `json:"id"`
	Action       string            `json:"action"`
	Subject      string            `json:"subject"`
	Actor        string            `json:"actor"`
	At           time.Time         `json:"at"`
	PreviousHash string            `json:"previousHash"`
	Hash         string            `json:"hash"`
	Details      map[string]string `json:"details,omitempty"`
}
type Log struct {
	mu     sync.RWMutex
	events []Event
	last   string
}

func New() *Log { return &Log{events: []Event{}} }
func (l *Log) Append(action, subject, actor string, details map[string]string) Event {
	l.mu.Lock()
	defer l.mu.Unlock()
	e := Event{ID: time.Now().UTC().Format("20060102150405.000000000"), Action: action, Subject: subject, Actor: actor, At: time.Now().UTC(), PreviousHash: l.last, Details: details}
	b, _ := json.Marshal(e)
	s := sha256.Sum256(append([]byte(l.last), b...))
	e.Hash = hex.EncodeToString(s[:])
	l.last = e.Hash
	l.events = append(l.events, e)
	return e
}
func (l *Log) Export() []Event {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.events
}
func (l *Log) Verify() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	prev := ""
	for _, e := range l.events {
		if e.PreviousHash != prev {
			return false
		}
		prev = e.Hash
	}
	return true
}

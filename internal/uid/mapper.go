package uid

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"sync"
)

type Mapper struct {
	mu     sync.RWMutex
	root   string
	values map[string]string
}

func New(root string) *Mapper { return &Mapper{root: root, values: make(map[string]string)} }
func (m *Mapper) Map(original string) string {
	m.mu.RLock()
	v := m.values[original]
	m.mu.RUnlock()
	if v != "" {
		return v
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if v = m.values[original]; v != "" {
		return v
	}
	sum := sha256.Sum256([]byte(original))
	n := binary.BigEndian.Uint64(sum[:8])
	v = fmt.Sprintf("%s.%d", m.root, n)
	m.values[original] = v
	return v
}
func (m *Mapper) Snapshot() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[string]string, len(m.values))
	for k, v := range m.values {
		out[k] = v
	}
	return out
}

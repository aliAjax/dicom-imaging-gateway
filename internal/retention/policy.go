package retention

import (
	"errors"
	"sort"
	"strings"
	"time"
)

type Policy struct {
	KeepFor             time.Duration
	Grace               time.Duration
	ProtectedModalities []string
}

func (p Policy) Normalize() (Policy, error) {
	if p.KeepFor <= 0 {
		return Policy{}, errors.New("retention duration must be positive")
	}
	if p.Grace < 0 {
		return Policy{}, errors.New("retention grace cannot be negative")
	}
	for i := range p.ProtectedModalities {
		p.ProtectedModalities[i] = strings.ToUpper(strings.TrimSpace(p.ProtectedModalities[i]))
	}
	sort.Strings(p.ProtectedModalities)
	return p, nil
}

package retention

import (
	"sort"
	"strings"
	"time"
)

type Object struct {
	ID         string
	Modality   string
	ReceivedAt time.Time
}

type Planner struct {
	Policy Policy
}

func (p Planner) Plan(now time.Time, objects []Object) ([]string, error) {
	policy, err := p.Policy.Normalize()
	if err != nil {
		return nil, err
	}
	protected := make(map[string]struct{}, len(policy.ProtectedModalities))
	for _, modality := range policy.ProtectedModalities {
		protected[modality] = struct{}{}
	}
	candidates := EligibleObjects(now, policy, objects)
	ids := make([]string, len(candidates))
	for _, object := range candidates {
		if _, ok := protected[strings.ToUpper(strings.TrimSpace(object.Modality))]; ok {
			continue
		}
		ids = append(ids, object.ID)
	}
	sort.Strings(ids)
	return ids, nil
}

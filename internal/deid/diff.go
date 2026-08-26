package deid

import (
	"example.com/dicom-gateway/internal/dicom"
	"fmt"
	"sort"
)

type Change struct {
	Tag    string `json:"tag"`
	Before string `json:"before"`
	After  string `json:"after"`
	Action Action `json:"action"`
}
type Diff struct {
	PolicyID    string   `json:"policyID"`
	FromVersion int      `json:"fromVersion"`
	ToVersion   int      `json:"toVersion"`
	Changes     []Change `json:"changes"`
}

func DiffPolicies(old, next Policy) Diff {
	d := Diff{PolicyID: next.ID, FromVersion: old.Version, ToVersion: next.Version}
	a := map[dicom.Tag]Rule{}
	for _, r := range old.Rules {
		a[r.Tag] = r
	}
	b := map[dicom.Tag]Rule{}
	for _, r := range next.Rules {
		b[r.Tag] = r
	}
	tags := []dicom.Tag{}
	for t := range a {
		tags = append(tags, t)
	}
	for t := range b {
		if _, ok := a[t]; !ok {
			tags = append(tags, t)
		}
	}
	sort.Slice(tags, func(i, j int) bool { return tags[i].String() < tags[j].String() })
	for _, t := range tags {
		if a[t] != b[t] {
			d.Changes = append(d.Changes, Change{Tag: t.String(), Before: fmt.Sprintf("%s:%s", a[t].Action, a[t].Value), After: fmt.Sprintf("%s:%s", b[t].Action, b[t].Value), Action: b[t].Action})
		}
	}
	return d
}

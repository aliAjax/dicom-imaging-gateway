package retention

import (
	"reflect"
	"testing"
	"time"
)

func TestNormalizeCopiesAndDeduplicatesModalities(t *testing.T) {
	input := []string{" mr ", "CT", "mr"}
	policy, err := (Policy{KeepFor: time.Hour, ProtectedModalities: input}).Normalize()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(input, []string{" mr ", "CT", "mr"}) {
		t.Fatalf("Normalize modified caller slice: %#v", input)
	}
	if !reflect.DeepEqual(policy.ProtectedModalities, []string{"CT", "MR"}) {
		t.Fatalf("normalized modalities = %#v", policy.ProtectedModalities)
	}
}

func TestEligibleObjectsDoesNotOverwriteInput(t *testing.T) {
	now := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	objects := []Object{{ID: "recent", ReceivedAt: now}, {ID: "old", ReceivedAt: now.Add(-48 * time.Hour)}}
	eligible := EligibleObjects(now, Policy{KeepFor: 24 * time.Hour}, objects)
	if !reflect.DeepEqual(eligible, []Object{{ID: "old", ReceivedAt: now.Add(-48 * time.Hour)}}) {
		t.Fatalf("eligible objects = %#v", eligible)
	}
	if objects[0].ID != "recent" || objects[1].ID != "old" {
		t.Fatalf("filter overwrote caller objects: %#v", objects)
	}
}

func TestPlannerReturnsCompactSortedIDs(t *testing.T) {
	now := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	old := now.Add(-48 * time.Hour)
	planner := Planner{Policy: Policy{KeepFor: 24 * time.Hour, ProtectedModalities: []string{"ct"}}}
	ids, err := planner.Plan(now, []Object{{ID: "z-study", Modality: "MR", ReceivedAt: old}, {ID: "a-study", Modality: " CT ", ReceivedAt: old}, {ID: "b-study", Modality: "US", ReceivedAt: old}})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(ids, []string{"b-study", "z-study"}) {
		t.Fatalf("planned IDs = %#v", ids)
	}
}

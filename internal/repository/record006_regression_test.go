package repository

import (
	"testing"

	"example.com/dicom-gateway/internal/dicom"
)

func TestQuerySnapshotOwnsResults(t *testing.T) {
	repo := New()
	if err := repo.PutInstance(dicom.Instance{UID: "ct-1", Metadata: dicom.Dataset{Modality: "CT"}}); err != nil {
		t.Fatalf("put CT instance: %v", err)
	}
	if err := repo.PutInstance(dicom.Instance{UID: "mr-1", Metadata: dicom.Dataset{Modality: "MR"}}); err != nil {
		t.Fatalf("put MR instance: %v", err)
	}
	first, _ := repo.QueryInstances(InstanceQuery{Modality: "CT"})
	if len(first) != 1 || first[0].UID != "ct-1" {
		t.Fatalf("unexpected first result: %#v", first)
	}
	second, _ := repo.QueryInstances(InstanceQuery{Modality: "MR"})
	if len(second) != 1 || second[0].UID != "mr-1" {
		t.Fatalf("unexpected second result: %#v", second)
	}
	if first[0].UID != "ct-1" {
		t.Fatalf("later query rewrote the earlier snapshot: %#v", first)
	}
}

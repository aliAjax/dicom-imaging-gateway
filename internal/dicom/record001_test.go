package dicom

import (
	"sync"
	"testing"
)

func TestAssociationSnapshotIsOwned(t *testing.T) {
	a := NewAssociation("assoc-1", "SOURCE", "ARCHIVE")
	a.AcceptedSyntaxes = []string{"1.2.840.10008.1.2.1", "1.2.840.10008.1.2"}
	snapshot := a.Snapshot()
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 4000; i++ {
			snapshot.AcceptedSyntaxes[0] = "changed"
		}
	}()
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 4000; i++ {
			_ = a.Snapshot().AcceptedSyntaxes[0]
		}
	}()
	close(start)
	wg.Wait()
	if got := a.Snapshot().AcceptedSyntaxes[0]; got != "1.2.840.10008.1.2.1" {
		t.Fatalf("snapshot mutation changed association syntax: %q", got)
	}
}

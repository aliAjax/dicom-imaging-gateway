package uid

import (
	"fmt"
	"sync"
	"testing"
)

func TestMapperSnapshotIsOwned(t *testing.T) {
	m := New("2.25.900")
	want := m.Map("1.2.840.1")
	snapshot := m.Snapshot()
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 2000; i++ {
			snapshot["1.2.840.1"] = fmt.Sprintf("changed-%d", i)
		}
	}()
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 2000; i++ {
			_ = m.Map("1.2.840.1")
			_ = m.Snapshot()
		}
	}()
	close(start)
	wg.Wait()
	if got := m.Map("1.2.840.1"); got != want {
		t.Fatalf("snapshot mutation changed mapper value: got %q want %q", got, want)
	}
}

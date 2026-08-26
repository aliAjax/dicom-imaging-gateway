package audit

import (
	"fmt"
	"sync"
	"testing"
)

func TestExportSnapshotConcurrentIsolation(t *testing.T) {
	log := New()
	for i := 0; i < 32; i++ {
		log.Append("stored", fmt.Sprintf("instance-%d", i), "worker", nil)
	}
	left, right := log.Export(), log.Export()
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 4000; i++ {
			left[0].Action = fmt.Sprintf("mutated-%d", i)
		}
	}()
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 4000; i++ {
			_ = right[0].Action
		}
	}()
	close(start)
	wg.Wait()
	if log.Export()[0].Action != "stored" {
		t.Fatal("caller mutation escaped into the audit log")
	}
}

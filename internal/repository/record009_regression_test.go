package repository

import (
	"fmt"
	"sync"
	"testing"

	"example.com/dicom-gateway/internal/dicom"
)

func TestInstanceSnapshotConcurrentIsolation(t *testing.T) {
	repo := New()
	for i := 0; i < 32; i++ {
		if err := repo.PutInstance(dicom.Instance{UID: fmt.Sprintf("instance-%d", i), Status: "stored"}); err != nil {
			t.Fatalf("seed instance: %v", err)
		}
	}
	left, right := repo.ListInstances(), repo.ListInstances()
	runSnapshotRace(t, func(i int) { left[0].Status = fmt.Sprintf("changed-%d", i) }, func() { _ = right[0].Status })
	if repo.ListInstances()[0].Status != "stored" {
		t.Fatal("caller mutation escaped into repository instances")
	}
}

func TestJobSnapshotConcurrentIsolation(t *testing.T) {
	repo := New()
	for i := 0; i < 32; i++ {
		repo.PutJob(dicom.RouteJob{ID: fmt.Sprintf("job-%d", i), Status: "queued"})
	}
	left, right := repo.Jobs(), repo.Jobs()
	runSnapshotRace(t, func(i int) { left[0].Status = fmt.Sprintf("changed-%d", i) }, func() { _ = right[0].Status })
	if repo.Jobs()[0].Status != "queued" {
		t.Fatal("caller mutation escaped into repository jobs")
	}
}

func runSnapshotRace(t *testing.T, write func(int), read func()) {
	t.Helper()
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 4000; i++ {
			write(i)
		}
	}()
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 4000; i++ {
			read()
		}
	}()
	close(start)
	wg.Wait()
}

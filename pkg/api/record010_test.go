package api

import (
	"example.com/dicom-gateway/internal/dicom"
	"testing"
)

func TestPageResponseOwnsItems(t *testing.T) {
	items := []dicom.Instance{{UID: "1.2.3"}}
	response := NewPageResponse(items, "next")
	items[0].UID = "corrupted"
	if response.Items[0].UID != "1.2.3" {
		t.Fatalf("page response changed through caller slice: %q", response.Items[0].UID)
	}
}

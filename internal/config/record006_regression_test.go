package config

import "testing"

func TestLoadStateIsRequestLocal(t *testing.T) {
	t.Setenv("DICOM_HTTP_ADDR", ":19061")
	first, err := Load()
	if err != nil {
		t.Fatalf("load first config: %v", err)
	}
	if first.HTTPAddr != ":19061" {
		t.Fatalf("first address = %q", first.HTTPAddr)
	}
	t.Setenv("DICOM_HTTP_ADDR", "")
	second, err := Load()
	if err != nil {
		t.Fatalf("load second config: %v", err)
	}
	if second.HTTPAddr != ":8080" {
		t.Fatalf("second load inherited prior address: %q", second.HTTPAddr)
	}
}

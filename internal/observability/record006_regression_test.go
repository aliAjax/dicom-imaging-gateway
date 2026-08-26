package observability

import "testing"

func TestJSONFailureDoesNotReusePayload(t *testing.T) {
	first := JSON(map[string]string{"request": "first"})
	if first == "" {
		t.Fatal("first payload was empty")
	}
	second := JSON(make(chan int))
	if second != "" {
		t.Fatalf("failed encoding reused prior payload: %q", second)
	}
}

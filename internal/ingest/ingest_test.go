package ingest

import (
	"context"
	"errors"
	"testing"
)

func TestDecodeFramesPreservesFrameOffset(t *testing.T) {
	frames, err := DecodeFrames([]byte{2, 'A', 'B', 4, 'C'})
	if len(frames) != 1 {
		t.Fatalf("decoded %d complete frames, want 1", len(frames))
	}
	var frameErr *FrameError
	if !errors.Is(err, ErrTruncatedFrame) || !errors.As(err, &frameErr) {
		t.Fatalf("error %v does not preserve frame error chain", err)
	}
	if frameErr.Index != 1 || frameErr.Offset != 3 {
		t.Fatalf("frame location = (%d, %d), want (1, 3)", frameErr.Index, frameErr.Offset)
	}
}

func TestBatchPreservesInputErrorChain(t *testing.T) {
	result := (Processor{}).Process(context.Background(), [][]byte{{1, 'A'}, {3, 'B'}, {1, 'C'}})
	if len(result.Items) != 3 {
		t.Fatalf("result slots = %d, want 3", len(result.Items))
	}
	if !errors.Is(result.Items[1].Err, ErrTruncatedFrame) {
		t.Fatalf("input error chain does not retain truncation: %v", result.Items[1].Err)
	}
	if len(result.Items[2].Frames) != 1 || string(result.Items[2].Frames[0]) != "C" {
		t.Fatalf("later valid instance was not retained: %#v", result.Items[2])
	}
}

func TestBatchErrorIncludesEveryFailure(t *testing.T) {
	first := errors.New("first instance")
	second := errors.New("second instance")
	err := (BatchResult{Items: []ItemResult{{Err: first}, {Err: second}}}).Error()
	if !errors.Is(err, first) || !errors.Is(err, second) {
		t.Fatalf("aggregate error %v does not retain both failures", err)
	}
}

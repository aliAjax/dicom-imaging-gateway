package dicom

import (
	"bytes"
	"errors"
	"testing"
)

type failingReader struct {
	err error
}

func (r failingReader) Read([]byte) (int, error) {
	return 0, r.err
}

func TestParsePreservesReaderFailureIdentity(t *testing.T) {
	want := errors.New("source read failed")
	_, err := (Parser{MaxElementBytes: 1024, MaxFileBytes: 1024}).Parse(failingReader{err: want})
	if !errors.Is(err, want) {
		t.Fatalf("reader error identity lost: %v", err)
	}
}

func TestParsePreservesMalformedClassification(t *testing.T) {
	data := append(make([]byte, 128), []byte("DICM")...)
	data = append(data, []byte{0x10, 0x00, 0x10, 0x00}...)
	_, err := (Parser{MaxElementBytes: 1024, MaxFileBytes: 1024}).Parse(bytes.NewReader(data))
	var pe *ParseError
	if !errors.As(err, &pe) {
		t.Fatalf("malformed ParseError identity lost: %v", err)
	}
	if pe.Kind != ErrMalformed || pe.Offset != 132 {
		t.Fatalf("malformed classification changed: %#v", pe)
	}
}

func TestParsePreservesSizeClassification(t *testing.T) {
	data := append(make([]byte, 128), []byte("DICM")...)
	data = append(data, 0)
	_, err := (Parser{MaxElementBytes: 1024, MaxFileBytes: 132}).Parse(bytes.NewReader(data))
	var pe *ParseError
	if !errors.As(err, &pe) {
		t.Fatalf("too-large ParseError identity lost: %v", err)
	}
	if pe.Kind != ErrTooLarge {
		t.Fatalf("size classification changed: %#v", pe)
	}
}

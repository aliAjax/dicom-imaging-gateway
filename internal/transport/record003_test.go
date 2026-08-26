package transport

import (
	"errors"
	"example.com/dicom-gateway/internal/dicom"
	"net/http"
	"testing"
)

type opaqueError struct {
	err error
}

func (opaqueError) Error() string {
	return "upload rejected"
}

func (e opaqueError) Unwrap() error {
	return e.err
}

func TestErrorMapperHandlesWrappedParseError(t *testing.T) {
	err := opaqueError{err: &dicom.ParseError{Kind: dicom.ErrTooLarge, Offset: 132, Message: "element exceeds limit"}}
	if got := (ErrorMapper{}).Status(err); got != http.StatusRequestEntityTooLarge {
		t.Fatalf("wrapped too-large error mapped to %d", got)
	}
}

func TestErrorMapperKeepsUnknownFailureInternal(t *testing.T) {
	if got := (ErrorMapper{}).Status(errors.New("storage unavailable")); got != http.StatusInternalServerError {
		t.Fatalf("unknown error mapped to %d", got)
	}
}

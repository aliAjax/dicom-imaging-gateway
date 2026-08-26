package dicom

import "fmt"

type ErrorKind string

const (
	ErrMalformed   ErrorKind = "malformed"
	ErrTooLarge    ErrorKind = "too_large"
	ErrUnsupported ErrorKind = "unsupported"
)

type ParseError struct {
	Kind    ErrorKind
	Offset  int
	Message string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("dicom %s at offset %d: %s", e.Kind, e.Offset, e.Message)
}
func malformed(off int, msg string) error {
	return fmt.Errorf("malformed DICOM input: %w", &ParseError{Kind: ErrMalformed, Offset: off, Message: msg})
}
func tooLarge(off int, msg string) error {
	return fmt.Errorf("oversized DICOM input: %w", &ParseError{Kind: ErrTooLarge, Offset: off, Message: msg})
}

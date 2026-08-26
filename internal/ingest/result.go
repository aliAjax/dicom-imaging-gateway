package ingest

import (
	"errors"
	"fmt"
	"strings"
)

type ItemResult struct {
	Frames [][]byte
	Err    error
}

type BatchResult struct {
	Items []ItemResult
}

// BatchError aggregates every failed input in a batch so the summary log
// never collapses to just the first failure. It preserves the underlying
// error chains so callers can detect specific failure kinds (for example
// ErrTruncatedFrame) and extract structured details (for example *FrameError)
// across the whole batch.
type BatchError struct {
	Errors []error
}

func (e *BatchError) Error() string {
	msgs := make([]string, len(e.Errors))
	for i, err := range e.Errors {
		msgs[i] = err.Error()
	}
	if len(msgs) == 1 {
		return fmt.Sprintf("batch failed: %s", msgs[0])
	}
	return fmt.Sprintf("batch failed (%d inputs): %s", len(msgs), strings.Join(msgs, "; "))
}

// Unwrap returns the per-input errors so errors.Is/errors.As traverse them,
// and multi-error unwrappers can enumerate every failure in the batch.
func (e *BatchError) Unwrap() []error { return e.Errors }

func (e *BatchError) Is(target error) bool {
	for _, err := range e.Errors {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

func (e *BatchError) As(target any) bool {
	for _, err := range e.Errors {
		if errors.As(err, target) {
			return true
		}
	}
	return false
}

// Error returns a single error that aggregates every failed input. A nil
// error means the whole batch succeeded.
func (r BatchResult) Error() error {
	var errs []error
	for _, item := range r.Items {
		if item.Err != nil {
			errs = append(errs, item.Err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return &BatchError{Errors: errs}
}

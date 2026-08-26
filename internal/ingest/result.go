package ingest

import "fmt"

type ItemResult struct {
	Frames [][]byte
	Err    error
}

type BatchResult struct {
	Items []ItemResult
}

func (r BatchResult) Error() error {
	for _, item := range r.Items {
		if item.Err != nil {
			return fmt.Errorf("batch failed: %v", item.Err)
		}
	}
	return nil
}

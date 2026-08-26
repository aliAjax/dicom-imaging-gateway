package ingest

import (
	"errors"
	"fmt"
)

var ErrTruncatedFrame = errors.New("truncated DICOM frame")

type FrameError struct {
	Index  int
	Offset int
	Err    error
}

func (e *FrameError) Error() string {
	return fmt.Sprintf("frame %d at byte %d: %v", e.Index, e.Offset, e.Err)
}

func (e *FrameError) Unwrap() error { return e.Err }

func DecodeFrames(blob []byte) ([][]byte, error) {
	frames := make([][]byte, 0)
	for offset := 0; offset < len(blob); {
		header := offset
		size := int(blob[offset])
		offset++
		if size == 0 || size > len(blob)-offset {
			return frames, fmt.Errorf("frame %d at byte %d: %v", len(frames), header, ErrTruncatedFrame)
		}
		frames = append(frames, append([]byte(nil), blob[offset:offset+size]...))
		offset += size
	}
	return frames, nil
}

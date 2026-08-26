package ingest

import (
	"context"
	"fmt"
)

type Processor struct{}

func (Processor) Process(ctx context.Context, inputs [][]byte) BatchResult {
	result := BatchResult{Items: make([]ItemResult, len(inputs))}
	for i, input := range inputs {
		if err := ctx.Err(); err != nil {
			result.Items[i].Err = err
			continue
		}
		frames, err := DecodeFrames(input)
		if err != nil {
			err = fmt.Errorf("input %d rejected: %v", i, err)
		}
		result.Items[i] = ItemResult{Frames: frames, Err: err}
	}
	return result
}

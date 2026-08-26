package dicom

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

const (
	ImplicitVRLittleEndian = "1.2.840.10008.1.2"
	ExplicitVRLittleEndian = "1.2.840.10008.1.2.1"
	RLELossless            = "1.2.840.10008.1.2.5"
	JPEGLSLossless         = "1.2.840.10008.1.2.4.80"
)

type Codec interface {
	TransferSyntax() string
	Decode(context.Context, []byte) ([]byte, error)
	Encode(context.Context, []byte) ([]byte, error)
}
type CodecRegistry struct {
	mu     sync.RWMutex
	codecs map[string]Codec
}

func NewCodecRegistry() *CodecRegistry {
	r := &CodecRegistry{codecs: map[string]Codec{}}
	r.Register(PassthroughCodec{ExplicitVRLittleEndian})
	r.Register(PassthroughCodec{ImplicitVRLittleEndian})
	return r
}
func (r *CodecRegistry) Register(c Codec) error {
	if c == nil || c.TransferSyntax() == "" {
		return errors.New("codec transfer syntax required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.codecs[c.TransferSyntax()]; ok {
		return fmt.Errorf("codec already registered: %s", c.TransferSyntax())
	}
	r.codecs[c.TransferSyntax()] = c
	return nil
}
func (r *CodecRegistry) Get(syntax string) (Codec, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.codecs[syntax]
	return c, ok
}
func (r *CodecRegistry) Transcode(ctx context.Context, source, target string, pixels []byte) ([]byte, error) {
	src, ok := r.Get(source)
	if !ok {
		return nil, fmt.Errorf("source codec unavailable: %s", source)
	}
	dst, ok := r.Get(target)
	if !ok {
		return nil, fmt.Errorf("target codec unavailable: %s", target)
	}
	raw, err := src.Decode(ctx, pixels)
	if err != nil {
		return nil, err
	}
	return dst.Encode(ctx, raw)
}

type PassthroughCodec struct{ Syntax string }

func (c PassthroughCodec) TransferSyntax() string { return c.Syntax }
func (c PassthroughCodec) Decode(ctx context.Context, b []byte) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	return append([]byte(nil), b...), nil
}
func (c PassthroughCodec) Encode(ctx context.Context, b []byte) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	return append([]byte(nil), b...), nil
}

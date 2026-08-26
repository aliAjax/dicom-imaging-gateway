package dicom

import (
	"context"
	"testing"
)

type record004Codec struct {
	syntax string
}

func (c *record004Codec) TransferSyntax() string { return c.syntax }
func (c *record004Codec) Decode(context.Context, []byte) ([]byte, error) {
	return []byte(c.syntax), nil
}
func (c *record004Codec) Encode(context.Context, []byte) ([]byte, error) {
	return []byte(c.syntax), nil
}

func TestRegistryRejectsTypedNilCodec(t *testing.T) {
	registry := &CodecRegistry{codecs: map[string]Codec{}}
	var codec *record004Codec
	if err := registry.Register(codec); err == nil {
		t.Fatal("typed nil codec was registered")
	}
}

func TestTranscodeMissingCodecReturnsError(t *testing.T) {
	registry := &CodecRegistry{codecs: map[string]Codec{}}
	var codec *record004Codec
	registry.codecs["typed-nil"] = codec
	registry.codecs[ExplicitVRLittleEndian] = PassthroughCodec{ExplicitVRLittleEndian}
	if _, err := registry.Transcode(context.Background(), "typed-nil", ExplicitVRLittleEndian, []byte("pixels")); err == nil {
		t.Fatal("typed nil source codec did not return an error")
	}
}

func TestNegotiationSkipsNilCapabilities(t *testing.T) {
	var missing *record004Codec
	valid := &record004Codec{syntax: ExplicitVRLittleEndian}
	got, err := SelectCodec([]Codec{missing, valid}, []string{ExplicitVRLittleEndian})
	if err != nil || got != valid {
		t.Fatalf("selected codec = %#v, err = %v", got, err)
	}
}

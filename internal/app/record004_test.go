package app

import (
	"bytes"
	"context"
	"example.com/dicom-gateway/internal/dicom"
	"testing"
)

type record004NilCodec struct{ marker *int }

func (c *record004NilCodec) TransferSyntax() string { return dicom.ExplicitVRLittleEndian }
func (c *record004NilCodec) Decode(context.Context, []byte) ([]byte, error) {
	_ = *c.marker
	return nil, nil
}
func (c *record004NilCodec) Encode(context.Context, []byte) ([]byte, error) { return nil, nil }

func TestValidationStopsAfterCodecFailure(t *testing.T) {
	service := &Service{Parser: dicom.Parser{MaxElementBytes: 1024, MaxFileBytes: 4096}}
	var codec *record004NilCodec
	if _, err := service.ValidateWithCodec(context.Background(), bytes.NewReader(nil), codec); err == nil {
		t.Fatal("typed nil codec did not stop validation")
	}
}

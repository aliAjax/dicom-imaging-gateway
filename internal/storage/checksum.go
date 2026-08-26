package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
)

func Digest(r io.Reader) (string, int64, error) {
	h := sha256.New()
	n, err := io.Copy(h, r)
	if err != nil {
		return "", n, err
	}
	return hex.EncodeToString(h.Sum(nil)), n, nil
}
func Verify(expected string, r io.Reader) error {
	got, _, err := Digest(r)
	if err != nil {
		return err
	}
	if got != expected {
		return io.ErrUnexpectedEOF
	}
	return nil
}

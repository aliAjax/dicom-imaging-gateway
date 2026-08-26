package dicom

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"hash"
)

type ChunkVerifier struct {
	expected  int64
	received  int64
	max       int64
	hasher    hash.Hash
	lastIndex int
}

func NewChunkVerifier(expected, max int64) *ChunkVerifier {
	return &ChunkVerifier{expected: expected, max: max, hasher: sha256.New(), lastIndex: -1}
}
func (v *ChunkVerifier) Add(index int, data []byte) error {
	if index != v.lastIndex+1 {
		return errors.New("pixel chunks must arrive in order")
	}
	if v.received+int64(len(data)) > v.max {
		return errors.New("pixel object exceeds configured limit")
	}
	v.hasher.Write(data)
	v.received += int64(len(data))
	v.lastIndex = index
	return nil
}
func (v *ChunkVerifier) Finish() (string, error) {
	if v.expected >= 0 && v.received != v.expected {
		return "", errors.New("pixel object length mismatch")
	}
	return hex.EncodeToString(v.hasher.Sum(nil)), nil
}
func ReassembleFragments(fragments [][]byte, max int64) ([]byte, error) {
	var total int64
	for _, f := range fragments {
		total += int64(len(f))
		if total > max {
			return nil, errors.New("fragment total exceeds limit")
		}
	}
	out := make([]byte, 0, total)
	for _, f := range fragments {
		out = append(out, f...)
	}
	return out, nil
}

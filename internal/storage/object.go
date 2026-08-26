package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type ObjectStore interface {
	Put(context.Context, string, io.Reader) (string, error)
	Get(context.Context, string) (io.ReadCloser, error)
	Delete(context.Context, string) error
}
type LocalStore struct{ Root string }

func NewLocal(root string) (*LocalStore, error) {
	if err := os.MkdirAll(root, 0750); err != nil {
		return nil, err
	}
	return &LocalStore{Root: root}, nil
}
func (s *LocalStore) Put(ctx context.Context, key string, r io.Reader) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	path := filepath.Join(s.Root, filepath.Clean("/"+key))
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return "", err
	}
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err = io.Copy(io.MultiWriter(f, h), r); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
func (s *LocalStore) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	return os.Open(filepath.Join(s.Root, filepath.Clean("/"+key)))
}
func (s *LocalStore) Delete(ctx context.Context, key string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return os.Remove(filepath.Join(s.Root, filepath.Clean("/"+key)))
}
func Key(uid string) string                       { return fmt.Sprintf("instances/%s.dcm", uid) }
func Expired(t time.Time, ttl time.Duration) bool { return time.Since(t) > ttl }

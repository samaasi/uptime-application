package storage

import (
	"context"
	"io"
	"time"
)

// Driver defines the interface for interacting with any storage backend.
type Driver interface {
	Upload(ctx context.Context, key string, data io.Reader, mimeType string) (string, error)
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	GetName() string
	GenerateSignedURL(ctx context.Context, key string, operation string, expires time.Duration) (string, error)
}

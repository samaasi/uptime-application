package database

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Client defines the interface for our PostgreSQL client.
type Client interface {
	DB() *gorm.DB
	WithContext(ctx context.Context) *gorm.DB
	Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error
	HealthCheck(ctx context.Context) error
	DebugDbInfo(ctx context.Context)
	Close() error
}

// CacheClient defines the interface for our Redis client.
type CacheClient interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, exp time.Duration) error
	Update(ctx context.Context, key string, value []byte) error
	Delete(ctx context.Context, key string) error
	Increment(ctx context.Context, key string) (int64, error)
	Decrement(ctx context.Context, key string) (int64, error)
	HealthCheck(ctx context.Context) error
	Close() error
}

// Row represents a single row result
type Row interface {
	Scan(dest ...interface{}) error
}

// SQLObjectToCreate represents a SQL statement and its description.
type SQLObjectToCreate struct {
	Query            string
	Description      string
	CheckExistsQuery string
}

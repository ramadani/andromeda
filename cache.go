package andromeda

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrCacheNotFound for error cache not found
	ErrCacheNotFound = errors.New("cache not found")
)

type Cache interface {
	IncrBy(ctx context.Context, key string, value int64) (int64, error)
	DecrBy(ctx context.Context, key string, decrement int64) (int64, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) (string, error)
	Get(ctx context.Context, key string) (string, error)
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	Exists(ctx context.Context, keys ...string) (int64, error)
	Del(ctx context.Context, keys ...string) (int64, error)
}

package andromeda

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrQuotaNotFound is error for quota not found
	ErrQuotaNotFound = errors.New("quota not found")
	// ErrQuotaLimitExceeded is error for quota exceeded
	ErrQuotaLimitExceeded = errors.New("quota limit exceeded")
)

// QuotaCache quota cache data
type QuotaCache struct {
	Key        string
	Expiration time.Duration
}

// AddQuotaUsage is a contract to add quota usage
type AddQuotaUsage interface {
	Do(ctx context.Context, id string, usage int64, data interface{}) (interface{}, error)
}

// ReduceQuotaUsage is a contract to reduce quota usage
type ReduceQuotaUsage interface {
	Do(ctx context.Context, id string, usage int64, data interface{}) (interface{}, error)
}

// XSetNXQuota is a contract to check exists or set if not exists for quota
type XSetNXQuota interface {
	Do(ctx context.Context, id string, data interface{}) error
}

// GetQuotaUsage is a contract to get quota usage
type GetQuotaUsage interface {
	Do(ctx context.Context, id string) (int64, error)
}

// GetQuotaCache is a contract to get quota cache from a given data
type GetQuotaCache interface {
	Do(ctx context.Context, id string, data interface{}) (*QuotaCache, error)
}

// GetQuota is a contract to get quota limit or usage
type GetQuota interface {
	Do(ctx context.Context, id string, data interface{}) (int64, error)
}

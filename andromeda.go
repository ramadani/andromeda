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

// UpdateQuotaUsage is a contract to update quota usage
// example: add quota usage or reduce quota usage
type UpdateQuotaUsage interface {
	Do(ctx context.Context, id string, value int64, data interface{}) (interface{}, error)
}

// GetQuotaCache is a contract to get quota cache from a given data
type GetQuotaCache interface {
	Do(ctx context.Context, id string, data interface{}) (*QuotaCache, error)
}

// XSetNXQuota is a contract to check exists or set if not exists for quota
type XSetNXQuota interface {
	Do(ctx context.Context, id string, data interface{}) error
}

// GetQuota is a contract to get quota limit or usage
type GetQuota interface {
	Do(ctx context.Context, id string, data interface{}) (int64, error)
}

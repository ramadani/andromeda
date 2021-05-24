package andromeda

import (
	"context"
	"time"
)

// QuotaRequest is a model for quota request
type QuotaRequest struct {
	QuotaID string
	Data    interface{}
}

// QuotaUsageRequest is a model for quota usage request
type QuotaUsageRequest struct {
	QuotaID string
	Usage   int64
	Data    interface{}
}

// QuotaCacheParams is a model for quota cache parameters
type QuotaCacheParams struct {
	Key        string
	Expiration time.Duration
}

// UpdateQuotaUsage is a contract to update quota usage
// example: add quota usage or reduce quota usage
type UpdateQuotaUsage interface {
	Do(ctx context.Context, req *QuotaUsageRequest) (interface{}, error)
}

// GetQuota is a contract to get quota limit or usage
type GetQuota interface {
	Do(ctx context.Context, req *QuotaRequest) (int64, error)
}

// XSetNXQuota is a contract to check exists or set if not exists for quota
type XSetNXQuota interface {
	Do(ctx context.Context, req *QuotaRequest) error
}

// GetQuotaCacheParams is a contract to get quota cache parameter from a given data
type GetQuotaCacheParams interface {
	Do(ctx context.Context, req *QuotaRequest) (*QuotaCacheParams, error)
}

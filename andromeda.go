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

// AddQuotaUsageConfig .
type AddQuotaUsageConfig struct {
	Cache               Cache
	GetQuotaLimit       GetQuota
	GetQuotaUsage       GetQuota
	GetQuotaCacheParams GetQuotaCacheParams
	Next                UpdateQuotaUsage
	Option              AddUsageOption
	LockInGetQuotaUsage time.Duration
}

// ReduceQuotaUsageConfig .
type ReduceQuotaUsageConfig struct {
	Cache               Cache
	GetQuotaUsage       GetQuota
	GetQuotaCacheParams GetQuotaCacheParams
	Next                UpdateQuotaUsage
	Option              ReduceUsageOption
	LockInGetQuotaUsage time.Duration
}

// AddQuotaUsage .
func AddQuotaUsage(conf AddQuotaUsageConfig) UpdateQuotaUsage {
	if conf.Next == nil {
		conf.Next = NopUpdateQuotaUsage()
	}

	addQuotaUsage := NewAddQuotaUsage(conf.Cache, conf.GetQuotaCacheParams, conf.GetQuotaLimit, conf.Next, conf.Option)

	if conf.GetQuotaUsage != nil && conf.LockInGetQuotaUsage.Milliseconds() > 0 {
		xSetNXQuotaUsage := NewXSetNXQuota(conf.Cache, conf.GetQuotaCacheParams, conf.GetQuotaUsage, conf.LockInGetQuotaUsage)
		addQuotaUsage = NewUpdateQuotaUsageMiddleware(NewXSetNXQuotaUsage(xSetNXQuotaUsage), addQuotaUsage)
	}

	return addQuotaUsage
}

// ReduceQuotaUsage .
func ReduceQuotaUsage(conf ReduceQuotaUsageConfig) UpdateQuotaUsage {
	if conf.Next == nil {
		conf.Next = NopUpdateQuotaUsage()
	}

	reduceQuotaUsage := NewReduceQuotaUsage(conf.Cache, conf.GetQuotaCacheParams, conf.Next, conf.Option)

	if conf.GetQuotaUsage != nil && conf.LockInGetQuotaUsage.Milliseconds() > 0 {
		xSetNXQuotaUsage := NewXSetNXQuota(conf.Cache, conf.GetQuotaCacheParams, conf.GetQuotaUsage, conf.LockInGetQuotaUsage)
		reduceQuotaUsage = NewUpdateQuotaUsageMiddleware(NewXSetNXQuotaUsage(xSetNXQuotaUsage), reduceQuotaUsage)
	}

	return reduceQuotaUsage
}

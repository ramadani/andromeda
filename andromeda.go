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

// UpdateQuotaUsageListener listen on success or error when updating quota usage
type UpdateQuotaUsageListener interface {
	OnSuccess(ctx context.Context, req *QuotaUsageRequest, updatedUsage int64)
	OnError(ctx context.Context, req *QuotaUsageRequest, err error)
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
	Next                UpdateQuotaUsage
	Cache               Cache
	GetQuotaCacheParams GetQuotaCacheParams
	GetQuotaLimit       GetQuota
	GetQuotaUsage       GetQuota
	GetQuotaUsageConfig GetQuotaUsageConfig
	Option              AddUsageOption
}

// ReduceQuotaUsageConfig .
type ReduceQuotaUsageConfig struct {
	Next                UpdateQuotaUsage
	Cache               Cache
	GetQuotaCacheParams GetQuotaCacheParams
	GetQuotaUsage       GetQuota
	GetQuotaUsageConfig GetQuotaUsageConfig
	Option              ReduceUsageOption
}

// GetQuotaUsageConfig .
type GetQuotaUsageConfig struct {
	LockIn   time.Duration
	MaxRetry int
	RetryIn  time.Duration
}

// GetLockIn .
func (q GetQuotaUsageConfig) GetLockIn() time.Duration {
	if q.LockIn.Milliseconds() > 0 {
		return q.LockIn
	}
	return time.Second * 1
}

func (q GetQuotaUsageConfig) GetMaxRetry() int {
	if q.MaxRetry > 0 {
		return q.MaxRetry
	}
	return 1
}

// GetRetryIn .
func (q GetQuotaUsageConfig) GetRetryIn() time.Duration {
	if q.RetryIn.Milliseconds() > 0 {
		return q.RetryIn
	}
	return time.Millisecond * 50
}

// AddQuotaUsage .
func AddQuotaUsage(conf AddQuotaUsageConfig) UpdateQuotaUsage {
	if conf.Next == nil {
		conf.Next = NopUpdateQuotaUsage()
	}

	addQuotaUsage := NewAddQuotaUsage(conf.Cache, conf.GetQuotaCacheParams, conf.GetQuotaLimit, conf.Next, conf.Option)

	if conf.GetQuotaUsage != nil {
		getUsageConf := conf.GetQuotaUsageConfig
		xSetNXQuotaUsage := NewXSetNXQuota(conf.Cache, conf.GetQuotaCacheParams, conf.GetQuotaUsage, getUsageConf.GetLockIn())
		xSetNXQuotaUsage = NewRetryableXSetNXQuota(xSetNXQuotaUsage, getUsageConf.GetMaxRetry(), getUsageConf.GetRetryIn())
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

	if conf.GetQuotaUsage != nil {
		getUsageConf := conf.GetQuotaUsageConfig
		xSetNXQuotaUsage := NewXSetNXQuota(conf.Cache, conf.GetQuotaCacheParams, conf.GetQuotaUsage, getUsageConf.GetLockIn())
		xSetNXQuotaUsage = NewRetryableXSetNXQuota(xSetNXQuotaUsage, getUsageConf.GetMaxRetry(), getUsageConf.GetRetryIn())
		reduceQuotaUsage = NewUpdateQuotaUsageMiddleware(NewXSetNXQuotaUsage(xSetNXQuotaUsage), reduceQuotaUsage)
	}

	return reduceQuotaUsage
}

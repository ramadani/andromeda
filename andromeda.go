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

// GetQuotaKey is a contract to get quota key for the cache
type GetQuotaKey interface {
	Do(ctx context.Context, req *QuotaRequest) (string, error)
}

// GetQuotaExpiration is a contract to get quota expiration for the cache
type GetQuotaExpiration interface {
	Do(ctx context.Context, req *QuotaRequest) (time.Duration, error)
}

// XSetNXQuota is a contract to check exists or set if not exists for quota
type XSetNXQuota interface {
	Do(ctx context.Context, req *QuotaRequest) error
}

// AddQuotaUsageConfig .
type AddQuotaUsageConfig struct {
	Next                    UpdateQuotaUsage
	Cache                   Cache
	GetQuotaLimit           GetQuota
	GetQuotaUsage           GetQuota
	GetQuotaUsageKey        GetQuotaKey
	GetQuotaUsageExpiration GetQuotaExpiration
	GetQuotaUsageConfig     GetQuotaUsageConfig
	Option                  AddUsageOption
}

// ReduceQuotaUsageConfig .
type ReduceQuotaUsageConfig struct {
	Next                    UpdateQuotaUsage
	Cache                   Cache
	GetQuotaUsage           GetQuota
	GetQuotaUsageKey        GetQuotaKey
	GetQuotaUsageExpiration GetQuotaExpiration
	GetQuotaUsageConfig     GetQuotaUsageConfig
	Option                  ReduceUsageOption
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
	if conf.Cache == nil {
		panic("Cache is required")
	}
	if conf.GetQuotaLimit == nil {
		panic("GetQuotaLimit is required")
	}
	if conf.GetQuotaUsageKey == nil {
		panic("GetQuotaUsageKey is required")
	}

	if conf.Next == nil {
		conf.Next = NopUpdateQuotaUsage()
	}

	addQuotaUsage := NewAddQuotaUsage(conf.Cache, conf.GetQuotaUsageKey, conf.GetQuotaLimit, conf.Next, conf.Option)

	if conf.GetQuotaUsage != nil {
		if conf.GetQuotaUsageExpiration == nil {
			panic("GetQuotaUsageExpiration is required")
		}

		getUsageConf := conf.GetQuotaUsageConfig
		xSetNXQuotaUsage := NewXSetNXQuota(conf.Cache, conf.GetQuotaUsageKey, conf.GetQuotaUsageExpiration, conf.GetQuotaUsage, getUsageConf.GetLockIn())
		xSetNXQuotaUsage = NewRetryableXSetNXQuota(xSetNXQuotaUsage, getUsageConf.GetMaxRetry(), getUsageConf.GetRetryIn())
		addQuotaUsage = NewUpdateQuotaUsageMiddleware(NewXSetNXQuotaUsage(xSetNXQuotaUsage), addQuotaUsage)
	}

	return addQuotaUsage
}

// ReduceQuotaUsage .
func ReduceQuotaUsage(conf ReduceQuotaUsageConfig) UpdateQuotaUsage {
	if conf.Cache == nil {
		panic("Cache is required")
	}
	if conf.GetQuotaUsageKey == nil {
		panic("GetQuotaUsageKey is required")
	}

	if conf.Next == nil {
		conf.Next = NopUpdateQuotaUsage()
	}

	reduceQuotaUsage := NewReduceQuotaUsage(conf.Cache, conf.GetQuotaUsageKey, conf.Next, conf.Option)

	if conf.GetQuotaUsage != nil {
		if conf.GetQuotaUsageExpiration == nil {
			panic("GetQuotaUsageExpiration is required")
		}

		getUsageConf := conf.GetQuotaUsageConfig
		xSetNXQuotaUsage := NewXSetNXQuota(conf.Cache, conf.GetQuotaUsageKey, conf.GetQuotaUsageExpiration, conf.GetQuotaUsage, getUsageConf.GetLockIn())
		xSetNXQuotaUsage = NewRetryableXSetNXQuota(xSetNXQuotaUsage, getUsageConf.GetMaxRetry(), getUsageConf.GetRetryIn())
		reduceQuotaUsage = NewUpdateQuotaUsageMiddleware(NewXSetNXQuotaUsage(xSetNXQuotaUsage), reduceQuotaUsage)
	}

	return reduceQuotaUsage
}

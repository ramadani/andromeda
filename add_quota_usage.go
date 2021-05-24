package andromeda

import (
	"context"
	"fmt"
)

// AddUsageOption .
type AddUsageOption struct {
	Reversible    bool
	ModifiedUsage int64
}

type addQuotaUsage struct {
	cache               Cache
	getQuotaCacheParams GetQuotaCacheParams
	getQuotaLimit       GetQuota
	next                UpdateQuotaUsage
	option              AddUsageOption
}

func (q *addQuotaUsage) Do(ctx context.Context, req *QuotaUsageRequest) (res interface{}, err error) {
	quotaReq := &QuotaRequest{QuotaID: req.QuotaID, Data: req.Data}

	cache, err := q.getQuotaCacheParams.Do(ctx, quotaReq)
	if err == ErrQuotaNotFound {
		return q.next.Do(ctx, req)
	} else if err != nil {
		return
	}

	usage := req.Usage
	if q.option.ModifiedUsage > 0 {
		usage = q.option.ModifiedUsage
	}

	limit, err := q.getQuotaLimit.Do(ctx, quotaReq)
	if err != nil {
		return
	}

	totalUsage, err := q.cache.IncrBy(ctx, cache.Key, usage)
	if err != nil {
		err = fmt.Errorf("%w: %v", ErrAddQuotaUsage, err)
		return
	}

	if totalUsage > limit {
		err = NewQuotaLimitExceededError(cache.Key, limit, totalUsage-usage)
		if er := q.reverseUsage(ctx, cache.Key, usage); er != nil {
			err = er
		}
		return
	}

	res, err = q.next.Do(ctx, req)

	if err != nil && q.option.Reversible {
		if er := q.reverseUsage(ctx, cache.Key, usage); er != nil {
			err = er
		}
	}
	return
}

func (q *addQuotaUsage) reverseUsage(ctx context.Context, key string, usage int64) error {
	if _, err := q.cache.DecrBy(ctx, key, usage); err != nil {
		return fmt.Errorf("%w: %v", ErrReduceQuotaUsage, err)
	}
	return nil
}

// NewAddQuotaUsage .
func NewAddQuotaUsage(
	cache Cache,
	getQuotaCacheParams GetQuotaCacheParams,
	getQuotaLimit GetQuota,
	next UpdateQuotaUsage,
	option AddUsageOption,
) UpdateQuotaUsage {
	return &addQuotaUsage{
		cache:               cache,
		getQuotaCacheParams: getQuotaCacheParams,
		getQuotaLimit:       getQuotaLimit,
		next:                next,
		option:              option,
	}
}

// NewQuotaLimitExceededError is a error helper for quota limit exceeded
func NewQuotaLimitExceededError(key string, limit, usage int64) error {
	return fmt.Errorf("%w: limit %d and usage %d for key %s", ErrQuotaLimitExceeded, limit, usage, key)
}

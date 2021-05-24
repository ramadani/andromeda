package andromeda

import (
	"context"
	"fmt"
)

// ReduceUsageOption .
type ReduceUsageOption struct {
	Reversible    bool
	ModifiedUsage int64
}

type reduceQuotaUsage struct {
	cache               Cache
	getQuotaCacheParams GetQuotaCacheParams
	next                UpdateQuotaUsage
	option              ReduceUsageOption
}

func (q *reduceQuotaUsage) Do(ctx context.Context, req *QuotaUsageRequest) (res interface{}, err error) {
	cache, err := q.getQuotaCacheParams.Do(ctx, &QuotaRequest{QuotaID: req.QuotaID, Data: req.Data})
	if err == ErrQuotaNotFound {
		return q.next.Do(ctx, req)
	} else if err != nil {
		return
	}

	usage := req.Usage
	if q.option.ModifiedUsage > 0 {
		usage = q.option.ModifiedUsage
	}

	totalUsage, err := q.cache.DecrBy(ctx, cache.Key, usage)
	if err != nil {
		err = fmt.Errorf("%w: %v", ErrReduceQuotaUsage, err)
		return
	}

	if totalUsage < 0 {
		err = NewInvalidMinQuotaUsageError(cache.Key, totalUsage)
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

func (q *reduceQuotaUsage) reverseUsage(ctx context.Context, key string, usage int64) error {
	if _, er := q.cache.IncrBy(ctx, key, usage); er != nil {
		return fmt.Errorf("%w: %v", ErrAddQuotaUsage, er)
	}
	return nil
}

// NewReduceQuotaUsage .
func NewReduceQuotaUsage(
	cache Cache,
	getQuotaCacheParams GetQuotaCacheParams,
	next UpdateQuotaUsage,
	option ReduceUsageOption,
) UpdateQuotaUsage {
	return &reduceQuotaUsage{
		cache:               cache,
		getQuotaCacheParams: getQuotaCacheParams,
		next:                next,
		option:              option,
	}
}

// NewInvalidMinQuotaUsageError is a error helper for invalid minimum quota usage
func NewInvalidMinQuotaUsageError(key string, usage int64) error {
	return fmt.Errorf("%w: usage %d for key %s", ErrInvalidMinQuotaUsage, usage, key)
}

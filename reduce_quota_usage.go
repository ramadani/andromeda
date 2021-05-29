package andromeda

import (
	"context"
	"fmt"
)

// ReduceUsageOption .
type ReduceUsageOption struct {
	ModifiedUsage int64
	Irreversible  bool // does not reverse when the next update quota usage has an error
	Listener      UpdateQuotaUsageListener
}

type reduceQuotaUsage struct {
	cache               Cache
	getQuotaCacheParams GetQuotaCacheParams
	next                UpdateQuotaUsage
	option              ReduceUsageOption
}

func (q *reduceQuotaUsage) Do(ctx context.Context, req *QuotaUsageRequest) (res interface{}, err error) {
	var totalUsage int64
	var isNextErr bool

	defer func() {
		if q.option.Listener != nil && !isNextErr {
			if err == nil {
				q.option.Listener.OnSuccess(ctx, req, totalUsage)
			} else {
				q.option.Listener.OnError(ctx, req, err)
			}
		}
	}()

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

	totalUsage, err = q.cache.DecrBy(ctx, cache.Key, usage)
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

	res, _err := q.next.Do(ctx, req)
	if _err != nil {
		isNextErr = true

		if !q.option.Irreversible {
			if er := q.reverseUsage(ctx, cache.Key, usage); er != nil {
				err, _err = er, er
				isNextErr = false
			}
		}
	}

	return res, _err
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

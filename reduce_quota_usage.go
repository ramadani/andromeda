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

func (q *reduceQuotaUsage) Do(ctx context.Context, id string, value int64, data interface{}) (res interface{}, err error) {
	cache, err := q.getQuotaCacheParams.Do(ctx, id, data)
	if err == ErrQuotaNotFound {
		return q.next.Do(ctx, id, value, data)
	} else if err != nil {
		return
	}

	usage := value
	if q.option.ModifiedUsage > 0 {
		usage = q.option.ModifiedUsage
	}

	if _, err = q.cache.DecrBy(ctx, cache.Key, usage); err != nil {
		err = fmt.Errorf("%w: %v", ErrReduceQuotaUsage, err)
		return
	}

	res, err = q.next.Do(ctx, id, value, data)

	if err != nil && q.option.Reversible {
		if _, er := q.cache.IncrBy(ctx, cache.Key, usage); er != nil {
			err = fmt.Errorf("%w: %v", ErrAddQuotaUsage, er)
		}
	}
	return
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

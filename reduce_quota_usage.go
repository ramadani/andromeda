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

	if _, err = q.cache.DecrBy(ctx, cache.Key, usage); err != nil {
		err = fmt.Errorf("%w: %v", ErrReduceQuotaUsage, err)
		return
	}

	res, err = q.next.Do(ctx, req)

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

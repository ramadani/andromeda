package andromeda

import (
	"context"
)

// ReduceUsageOption .
type ReduceUsageOption struct {
	Reversible    bool
	ModifiedUsage int64
}

type reduceQuotaUsage struct {
	cache         Cache
	getKeyLimit   GetQuotaCache
	next          UpdateQuotaUsage
	reversible    bool
	modifiedUsage int64
}

func (q *reduceQuotaUsage) Do(ctx context.Context, id string, value int64, data interface{}) (res interface{}, err error) {
	cache, err := q.getKeyLimit.Do(ctx, id, data)
	if err == ErrQuotaNotFound {
		return q.next.Do(ctx, id, value, data)
	} else if err != nil {
		return
	}

	usage := value
	if q.modifiedUsage > 0 {
		usage = q.modifiedUsage
	}

	if _, err = q.cache.DecrBy(ctx, cache.Key, usage); err != nil {
		return
	}

	defer func() {
		if err != nil && q.reversible {
			if _, er := q.cache.IncrBy(ctx, cache.Key, usage); er != nil {
				err = er
			}
		}
	}()

	res, err = q.next.Do(ctx, id, value, data)
	return
}

// NewReduceQuotaUsage .
func NewReduceQuotaUsage(
	cache Cache,
	getKeyLimit GetQuotaCache,
	next UpdateQuotaUsage,
	option ReduceUsageOption,
) UpdateQuotaUsage {
	return &reduceQuotaUsage{
		cache:         cache,
		getKeyLimit:   getKeyLimit,
		next:          next,
		reversible:    option.Reversible,
		modifiedUsage: option.ModifiedUsage,
	}
}

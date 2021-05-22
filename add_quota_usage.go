package andromeda

import (
	"context"
	"fmt"
)

// AddUsageOption .
type AddUsageOption struct {
	ModifiedUsage int64
}

type addQuotaUsage struct {
	cache         Cache
	getQuotaCache GetQuotaCache
	getQuotaLimit GetQuota
	next          AddQuotaUsage
	modifiedUsage int64
}

func (q *addQuotaUsage) Do(ctx context.Context, id string, usage int64, data interface{}) (res interface{}, err error) {
	cache, err := q.getQuotaCache.Do(ctx, id, data)
	if err == ErrQuotaNotFound {
		return q.next.Do(ctx, id, usage, data)
	} else if err != nil {
		return
	}

	quotaUsage := usage
	if q.modifiedUsage > 0 {
		quotaUsage = q.modifiedUsage
	}

	limit, err := q.getQuotaLimit.Do(ctx, id, data)
	if err != nil {
		return
	}

	totalUsage, err := q.cache.IncrBy(ctx, cache.Key, quotaUsage)
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			if _, er := q.cache.DecrBy(ctx, cache.Key, quotaUsage); er != nil {
				err = er
			}
		}
	}()

	if totalUsage > limit {
		err = NewQuotaLimitExceededError(cache.Key, limit, totalUsage-quotaUsage)
		return
	}

	res, err = q.next.Do(ctx, id, usage, data)
	return
}

// NewAddQuotaUsage .
func NewAddQuotaUsage(
	cache Cache,
	getQuotaCache GetQuotaCache,
	getQuotaLimit GetQuota,
	next AddQuotaUsage,
	option AddUsageOption,
) AddQuotaUsage {
	return &addQuotaUsage{
		cache:         cache,
		getQuotaCache: getQuotaCache,
		getQuotaLimit: getQuotaLimit,
		next:          next,
		modifiedUsage: option.ModifiedUsage,
	}
}

// NewQuotaLimitExceededError is a error helper for quota limit exceeded
func NewQuotaLimitExceededError(key string, limit, usage int64) error {
	return fmt.Errorf("%w: limit %d and usage %d for key %s", ErrQuotaLimitExceeded, limit, usage, key)
}

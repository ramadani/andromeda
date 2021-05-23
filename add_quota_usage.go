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
	cache               Cache
	getQuotaCacheParams GetQuotaCacheParams
	getQuotaLimit       GetQuota
	next                UpdateQuotaUsage
	modifiedUsage       int64
}

func (q *addQuotaUsage) Do(ctx context.Context, id string, value int64, data interface{}) (res interface{}, err error) {
	cache, err := q.getQuotaCacheParams.Do(ctx, id, data)
	if err == ErrQuotaNotFound {
		return q.next.Do(ctx, id, value, data)
	} else if err != nil {
		return
	}

	usage := value
	if q.modifiedUsage > 0 {
		usage = q.modifiedUsage
	}

	limit, err := q.getQuotaLimit.Do(ctx, id, data)
	if err != nil {
		return
	}

	totalUsage, err := q.cache.IncrBy(ctx, cache.Key, usage)
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			if _, er := q.cache.DecrBy(ctx, cache.Key, usage); er != nil {
				err = er
			}
		}
	}()

	if totalUsage > limit {
		err = NewQuotaLimitExceededError(cache.Key, limit, totalUsage-usage)
		return
	}

	res, err = q.next.Do(ctx, id, value, data)
	return
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
		modifiedUsage:       option.ModifiedUsage,
	}
}

// NewQuotaLimitExceededError is a error helper for quota limit exceeded
func NewQuotaLimitExceededError(key string, limit, usage int64) error {
	return fmt.Errorf("%w: limit %d and usage %d for key %s", ErrQuotaLimitExceeded, limit, usage, key)
}

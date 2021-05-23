package andromeda

import (
	"context"
	"fmt"
	"strconv"
)

type getCachedQuota struct {
	cache         Cache
	getQuotaCache GetQuotaCache
}

func (q *getCachedQuota) Do(ctx context.Context, id string, data interface{}) (int64, error) {
	cache, err := q.getQuotaCache.Do(ctx, id, data)
	if err != nil {
		return 0, err
	}

	val, err := q.cache.Get(ctx, cache.Key)
	if err == ErrCacheNotFound {
		return 0, fmt.Errorf("%w: key %s", ErrQuotaNotFound, cache.Key)
	} else if err != nil {
		return 0, err
	}

	res, _ := strconv.Atoi(val)

	return int64(res), nil
}

// NewGetCachedQuota .
func NewGetCachedQuota(cache Cache, getQuotaCache GetQuotaCache) GetQuota {
	return &getCachedQuota{cache: cache, getQuotaCache: getQuotaCache}
}

package andromeda

import (
	"context"
	"fmt"
	"strconv"
)

type getCachedQuota struct {
	cache       Cache
	getQuotaKey GetQuotaKey
}

func (q *getCachedQuota) Do(ctx context.Context, req *QuotaRequest) (int64, error) {
	key, err := q.getQuotaKey.Do(ctx, req)
	if err != nil {
		return 0, err
	}

	val, err := q.cache.Get(ctx, key)
	if err == ErrCacheNotFound {
		return 0, fmt.Errorf("%w: key %s", ErrQuotaNotFound, key)
	} else if err != nil {
		return 0, err
	}

	res, _ := strconv.Atoi(val)

	return int64(res), nil
}

// NewGetCachedQuota .
func NewGetCachedQuota(cache Cache, getQuotaKey GetQuotaKey) GetQuota {
	return &getCachedQuota{cache: cache, getQuotaKey: getQuotaKey}
}

package andromeda

import (
	"context"
	"fmt"
	"strconv"
)

type getCachedQuota struct {
	cache               Cache
	getQuotaCacheParams GetQuotaCacheParams
}

func (q *getCachedQuota) Do(ctx context.Context, req *QuotaRequest) (int64, error) {
	cache, err := q.getQuotaCacheParams.Do(ctx, req)
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
func NewGetCachedQuota(cache Cache, getQuotaCacheParams GetQuotaCacheParams) GetQuota {
	return &getCachedQuota{cache: cache, getQuotaCacheParams: getQuotaCacheParams}
}

package andromeda

import (
	"context"
	"fmt"
	"strconv"
)

type getQuotaUsage struct {
	cache Cache
}

func (q *getQuotaUsage) Do(ctx context.Context, id string) (int64, error) {
	val, err := q.cache.Get(ctx, id)
	if err == ErrCacheNotFound {
		return 0, fmt.Errorf("%w: key %s", ErrQuotaNotFound, id)
	} else if err != nil {
		return 0, err
	}

	res, _ := strconv.Atoi(val)

	return int64(res), nil
}

// NewGetQuotaUsage .
func NewGetQuotaUsage(cache Cache) GetQuotaUsage {
	return &getQuotaUsage{cache: cache}
}

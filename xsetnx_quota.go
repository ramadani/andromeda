package andromeda

import (
	"context"
	"fmt"
	"time"
)

type xSetNXQuota struct {
	cache         Cache
	getQuotaCache GetQuotaCache
	getQuota      GetQuota
	lockIn        time.Duration
}

func (q *xSetNXQuota) Do(ctx context.Context, id string, data interface{}) (err error) {
	cache, err := q.getQuotaCache.Do(ctx, id, data)
	if err != nil {
		return
	}

	exists, err := q.cache.Exists(ctx, cache.Key)
	if err != nil {
		return
	} else if exists == 1 {
		return
	}

	lockKey := fmt.Sprintf("%s-lock", cache.Key)
	succeedLock, err := q.cache.SetNX(ctx, lockKey, 1, q.lockIn)
	if err != nil {
		return
	} else if !succeedLock {
		err = fmt.Errorf("locked key %s", lockKey)
		return
	}

	defer func() {
		_, err = q.cache.Del(ctx, lockKey)
	}()

	val, err := q.getQuota.Do(ctx, id, data)
	if err != nil {
		return
	}

	_, err = q.cache.SetNX(ctx, cache.Key, val, cache.Expiration)
	return
}

// NewXSetNXQuota .
func NewXSetNXQuota(
	cache Cache,
	getQuotaCache GetQuotaCache,
	getQuota GetQuota,
	lockIn time.Duration,
) XSetNXQuota {
	return &xSetNXQuota{
		cache:         cache,
		getQuotaCache: getQuotaCache,
		getQuota:      getQuota,
		lockIn:        lockIn,
	}
}

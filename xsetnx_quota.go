package andromeda

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	// ErrLockedKey .
	ErrLockedKey = errors.New("locked key")
	// ErrMaxRetryExceeded .
	ErrMaxRetryExceeded = errors.New("max retry exceeded")
)

type xSetNXQuota struct {
	cache               Cache
	getQuotaCacheParams GetQuotaCacheParams
	getQuota            GetQuota
	lockIn              time.Duration
}

func (q *xSetNXQuota) Do(ctx context.Context, id string, data interface{}) (err error) {
	cache, err := q.getQuotaCacheParams.Do(ctx, id, data)
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
		err = fmt.Errorf("%w: %s", ErrLockedKey, lockKey)
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
	getQuotaCacheParams GetQuotaCacheParams,
	getQuota GetQuota,
	lockIn time.Duration,
) XSetNXQuota {
	return &xSetNXQuota{
		cache:               cache,
		getQuotaCacheParams: getQuotaCacheParams,
		getQuota:            getQuota,
		lockIn:              lockIn,
	}
}

type retryableXSetNXQuota struct {
	next     XSetNXQuota
	maxRetry int
	sleepIn  time.Duration
}

func (q *retryableXSetNXQuota) Do(ctx context.Context, id string, data interface{}) error {
	for i := 0; i < q.maxRetry; i++ {
		if err := q.next.Do(ctx, id, data); err == nil {
			return nil
		}

		if i+1 != q.maxRetry {
			time.Sleep(q.sleepIn)
		}
	}

	return ErrMaxRetryExceeded
}

// NewRetryableXSetNXQuota .
func NewRetryableXSetNXQuota(next XSetNXQuota, maxRetry int, sleepIn time.Duration) XSetNXQuota {
	return &retryableXSetNXQuota{
		next:     next,
		maxRetry: maxRetry,
		sleepIn:  sleepIn,
	}
}

package andromeda

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type xSetNXQuota struct {
	cache              Cache
	getQuotaKey        GetQuotaKey
	getQuotaExpiration GetQuotaExpiration
	getQuota           GetQuota
	lockIn             time.Duration
}

func (q *xSetNXQuota) Do(ctx context.Context, req *QuotaRequest) (err error) {
	key, err := q.getQuotaKey.Do(ctx, req)
	if errors.Is(err, ErrQuotaNotFound) {
		return nil
	} else if err != nil {
		return
	}

	exists, err := q.cache.Exists(ctx, key)
	if err != nil || exists == 1 {
		return
	}

	lockKey := fmt.Sprintf("%s-lock", key)
	succeedLock, err := q.cache.SetNX(ctx, lockKey, 1, q.lockIn)
	if err != nil {
		return
	} else if !succeedLock {
		err = fmt.Errorf("%w: %s", ErrLockedKey, lockKey)
		return
	}

	defer func() {
		if _, er := q.cache.Del(ctx, lockKey); er != nil {
			err = er
		}
	}()

	val, err := q.getQuota.Do(ctx, req)
	if err != nil {
		return
	}

	exp, err := q.getQuotaExpiration.Do(ctx, req)
	if err != nil {
		return
	}

	_, err = q.cache.SetNX(ctx, key, val, exp)
	return
}

// NewXSetNXQuota .
func NewXSetNXQuota(
	cache Cache,
	getQuotaKey GetQuotaKey,
	getQuotaExpiration GetQuotaExpiration,
	getQuota GetQuota,
	lockIn time.Duration,
) XSetNXQuota {
	return &xSetNXQuota{
		cache:              cache,
		getQuotaKey:        getQuotaKey,
		getQuotaExpiration: getQuotaExpiration,
		getQuota:           getQuota,
		lockIn:             lockIn,
	}
}

type retryableXSetNXQuota struct {
	next     XSetNXQuota
	maxRetry int
	retryIn  time.Duration
}

func (q *retryableXSetNXQuota) Do(ctx context.Context, req *QuotaRequest) error {
	var err error
	for i := 0; i < q.maxRetry; i++ {
		err = q.next.Do(ctx, req)
		if err == nil {
			return nil
		}

		if i+1 != q.maxRetry {
			time.Sleep(q.retryIn)
		}
	}

	return fmt.Errorf("%w: %q", ErrMaxRetryExceeded, err)
}

// NewRetryableXSetNXQuota .
func NewRetryableXSetNXQuota(next XSetNXQuota, maxRetry int, sleepIn time.Duration) XSetNXQuota {
	return &retryableXSetNXQuota{
		next:     next,
		maxRetry: maxRetry,
		retryIn:  sleepIn,
	}
}

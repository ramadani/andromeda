package andromeda

import (
	"context"
	"fmt"
)

// AddUsageOption .
type AddUsageOption struct {
	Reversible    bool
	ModifiedUsage int64
	Listener      UpdateQuotaUsageListener
}

type addQuotaUsage struct {
	cache               Cache
	getQuotaCacheParams GetQuotaCacheParams
	getQuotaLimit       GetQuota
	next                UpdateQuotaUsage
	option              AddUsageOption
}

func (q *addQuotaUsage) Do(ctx context.Context, req *QuotaUsageRequest) (res interface{}, err error) {
	var totalUsage int64
	var isNextErr bool

	defer func() {
		if q.option.Listener != nil && !isNextErr {
			if err == nil {
				q.option.Listener.OnSuccess(ctx, req, totalUsage)
			} else {
				q.option.Listener.OnError(ctx, req, err)
			}
		}
	}()

	quotaReq := &QuotaRequest{QuotaID: req.QuotaID, Data: req.Data}
	cache, err := q.getQuotaCacheParams.Do(ctx, quotaReq)
	if err == ErrQuotaNotFound {
		return q.next.Do(ctx, req)
	} else if err != nil {
		return
	}

	usage := req.Usage
	if q.option.ModifiedUsage > 0 {
		usage = q.option.ModifiedUsage
	}

	limit, err := q.getQuotaLimit.Do(ctx, quotaReq)
	if err != nil {
		return
	}

	totalUsage, err = q.cache.IncrBy(ctx, cache.Key, usage)
	if err != nil {
		err = fmt.Errorf("%w: %v", ErrAddQuotaUsage, err)
		return
	}

	if totalUsage > limit {
		err = NewQuotaLimitExceededError(cache.Key, limit, totalUsage-usage)
		if er := q.reverseUsage(ctx, cache.Key, usage); er != nil {
			err = er
		}
		return
	}

	res, _err := q.next.Do(ctx, req)
	if _err != nil {
		isNextErr = true

		if q.option.Reversible {
			if er := q.reverseUsage(ctx, cache.Key, usage); er != nil {
				err, _err = er, er
				isNextErr = false
			}
		}
	}

	return res, _err
}

func (q *addQuotaUsage) reverseUsage(ctx context.Context, key string, usage int64) error {
	if _, err := q.cache.DecrBy(ctx, key, usage); err != nil {
		return fmt.Errorf("%w: %v", ErrReduceQuotaUsage, err)
	}
	return nil
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
		option:              option,
	}
}

// NewQuotaLimitExceededError is a error helper for quota limit exceeded
func NewQuotaLimitExceededError(key string, limit, usage int64) error {
	return fmt.Errorf("%w: limit %d and usage %d for key %s", ErrQuotaLimitExceeded, limit, usage, key)
}

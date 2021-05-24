package andromeda

import "context"

type updateQuotaUsageMiddleware struct {
	prev, next UpdateQuotaUsage
}

func (q *updateQuotaUsageMiddleware) Do(ctx context.Context, req *QuotaUsageRequest) (interface{}, error) {
	res, err := q.prev.Do(ctx, req)
	if err != nil {
		return res, err
	}

	return q.next.Do(ctx, req)
}

// NewUpdateQuotaUsageMiddleware .
func NewUpdateQuotaUsageMiddleware(prev, next UpdateQuotaUsage) UpdateQuotaUsage {
	return &updateQuotaUsageMiddleware{
		prev: prev,
		next: next,
	}
}

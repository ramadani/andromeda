package andromeda

import "context"

type updateQuotaUsageMiddleware struct {
	prev, next UpdateQuotaUsage
}

func (q *updateQuotaUsageMiddleware) Do(ctx context.Context, id string, value int64, data interface{}) (interface{}, error) {
	res, err := q.prev.Do(ctx, id, value, data)
	if err != nil {
		return res, err
	}

	return q.next.Do(ctx, id, value, data)
}

// NewUpdateQuotaUsageMiddleware .
func NewUpdateQuotaUsageMiddleware(prev, next UpdateQuotaUsage) UpdateQuotaUsage {
	return &updateQuotaUsageMiddleware{
		prev: prev,
		next: next,
	}
}

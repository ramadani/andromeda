package andromeda

import "context"

type nopUpdateQuotaUsage struct{}

func (q *nopUpdateQuotaUsage) Do(ctx context.Context, req *QuotaUsageRequest) (interface{}, error) {
	return nil, nil
}

// NopUpdateQuotaUsage .
func NopUpdateQuotaUsage() UpdateQuotaUsage {
	return &nopUpdateQuotaUsage{}
}

type xSetNXQuotaUsage struct {
	xSetNXQuota XSetNXQuota
}

func (q *xSetNXQuotaUsage) Do(ctx context.Context, req *QuotaUsageRequest) (interface{}, error) {
	err := q.xSetNXQuota.Do(ctx, &QuotaRequest{QuotaID: req.QuotaID, Data: req.Data})

	return nil, err
}

// NewXSetNXQuotaUsage .
func NewXSetNXQuotaUsage(xSetNXQuota XSetNXQuota) UpdateQuotaUsage {
	return &xSetNXQuotaUsage{xSetNXQuota: xSetNXQuota}
}

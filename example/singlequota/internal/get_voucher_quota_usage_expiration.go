package internal

import (
	"context"
	"github.com/ramadani/andromeda"
	"time"
)

type getVoucherQuotaUsageExpiration struct{}

func (v *getVoucherQuotaUsageExpiration) Do(ctx context.Context, req *andromeda.QuotaRequest) (time.Duration, error) {
	return time.Hour * 5, nil
}

func NewGetVoucherQuotaUsageExpiration() andromeda.GetQuotaExpiration {
	return &getVoucherQuotaUsageExpiration{}
}

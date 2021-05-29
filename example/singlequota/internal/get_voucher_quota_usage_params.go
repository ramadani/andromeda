package internal

import (
	"context"
	"fmt"
	"github.com/ramadani/andromeda"
	"time"
)

type getVoucherQuotaUsageParams struct {
	keyFormat string
}

func (v *getVoucherQuotaUsageParams) Do(_ context.Context, req *andromeda.QuotaRequest) (*andromeda.QuotaCacheParams, error) {
	key := fmt.Sprintf(v.keyFormat, req.QuotaID)
	expiration := 5 * time.Minute

	return &andromeda.QuotaCacheParams{
		Key:        key,
		Expiration: expiration,
	}, nil
}

func NewGetVoucherQuotaParams(keyFormat string) andromeda.GetQuotaCacheParams {
	return &getVoucherQuotaUsageParams{keyFormat: keyFormat}
}

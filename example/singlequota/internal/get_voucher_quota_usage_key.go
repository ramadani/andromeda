package internal

import (
	"context"
	"fmt"
	"github.com/ramadani/andromeda"
)

type getVoucherQuotaUsageKey struct {
	keyFormat string
}

func (v *getVoucherQuotaUsageKey) Do(_ context.Context, req *andromeda.QuotaRequest) (string, error) {
	return fmt.Sprintf(v.keyFormat, req.QuotaID), nil
}

func NewGetVoucherQuotaUsageKey(keyFormat string) andromeda.GetQuotaKey {
	return &getVoucherQuotaUsageKey{keyFormat: keyFormat}
}

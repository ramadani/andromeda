package internal

import (
	"context"
	"github.com/ramadani/andromeda"
	"log"
)

type updateVoucherQuotaUsageListener struct{}

func (v *updateVoucherQuotaUsageListener) OnSuccess(ctx context.Context, req *andromeda.QuotaUsageRequest, updatedUsage int64) {
	log.Println("updated quota", updatedUsage)
}

func (v *updateVoucherQuotaUsageListener) OnError(ctx context.Context, req *andromeda.QuotaUsageRequest, err error) {
	log.Println("err", err)
}

func NewUpdateVoucherQuotaUsageListener() andromeda.UpdateQuotaUsageListener {
	return &updateVoucherQuotaUsageListener{}
}

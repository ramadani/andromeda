package internal

import (
	"context"
	"github.com/ramadani/andromeda"
)

type getVoucherQuotaUsage struct {
	voucherRepo VoucherRepository
}

func (v *getVoucherQuotaUsage) Do(ctx context.Context, req *andromeda.QuotaRequest) (int64, error) {
	voucher, err := v.voucherRepo.FindByID(ctx, req.QuotaID)
	if err != nil {
		return 0, err
	}

	return voucher.Usage, nil
}

func NewGetVoucherQuotaUsage(voucherRepo VoucherRepository) andromeda.GetQuota {
	return &getVoucherQuotaUsage{voucherRepo: voucherRepo}
}

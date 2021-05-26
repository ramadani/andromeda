package internal

import (
	"context"
	"github.com/ramadani/andromeda"
)

type getVoucherQuotaLimit struct {
	voucherRepo VoucherRepository
}

func (v *getVoucherQuotaLimit) Do(ctx context.Context, req *andromeda.QuotaRequest) (int64, error) {
	voucher, err := v.voucherRepo.FindByID(ctx, req.QuotaID)
	if err != nil {
		return 0, err
	}

	return voucher.Limit, nil
}

func NewGetVoucherQuotaLimit(voucherRepo VoucherRepository) andromeda.GetQuota {
	return &getVoucherQuotaLimit{voucherRepo: voucherRepo}
}

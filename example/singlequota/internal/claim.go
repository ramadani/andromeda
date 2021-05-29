package internal

import (
	"context"
	"github.com/ramadani/andromeda"
	uuid "github.com/satori/go.uuid"
	"time"
)

type ClaimVoucher interface {
	Do(ctx context.Context, code, userID string) (*History, error)
}

type claimVoucher struct {
	voucherRepo        VoucherRepository
	historyRepo        HistoryRepository
	addVoucherUsage    andromeda.UpdateQuotaUsage
	reduceVoucherUsage andromeda.UpdateQuotaUsage
}

func (c *claimVoucher) Do(ctx context.Context, code, userID string) (*History, error) {
	voucher, err := c.voucherRepo.FindByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	history := &History{
		ID:        uuid.NewV4().String(),
		VoucherID: voucher.ID,
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	usageReq := andromeda.QuotaUsageRequest{
		QuotaID: voucher.ID,
		Usage:   1,
	}

	if _, err = c.addVoucherUsage.Do(ctx, &usageReq); err != nil {
		return nil, err
	}

	if err = c.historyRepo.Create(ctx, history); err == nil {
		return history, nil
	}

	// has error, do reverse usage quota
	if _, er := c.reduceVoucherUsage.Do(ctx, &usageReq); er != nil {
		err = er
	}

	return nil, err
}

func NewClaimVoucher(
	voucherRepo VoucherRepository,
	historyRepo HistoryRepository,
	addVoucherUsage andromeda.UpdateQuotaUsage,
	reduceVoucherUsage andromeda.UpdateQuotaUsage,
) ClaimVoucher {
	return &claimVoucher{
		voucherRepo:        voucherRepo,
		historyRepo:        historyRepo,
		addVoucherUsage:    addVoucherUsage,
		reduceVoucherUsage: reduceVoucherUsage,
	}
}

type unClaimVoucher struct {
	voucherRepo        VoucherRepository
	historyRepo        HistoryRepository
	addVoucherUsage    andromeda.UpdateQuotaUsage
	reduceVoucherUsage andromeda.UpdateQuotaUsage
}

func (c *unClaimVoucher) Do(ctx context.Context, code, userID string) (*History, error) {
	voucher, err := c.voucherRepo.FindByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	history := &History{
		ID:        uuid.NewV4().String(),
		VoucherID: voucher.ID,
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	usageReq := andromeda.QuotaUsageRequest{
		QuotaID: voucher.ID,
		Usage:   1,
	}

	if _, err = c.reduceVoucherUsage.Do(ctx, &usageReq); err != nil {
		return nil, err
	}

	if err = c.historyRepo.Create(ctx, history); err == nil {
		return history, nil
	}

	// has error, do reverse usage quota
	if _, er := c.addVoucherUsage.Do(ctx, &usageReq); er != nil {
		err = er
	}

	return nil, err
}

func NewUnClaimVoucher(
	voucherRepo VoucherRepository,
	historyRepo HistoryRepository,
	addVoucherUsage andromeda.UpdateQuotaUsage,
	reduceVoucherUsage andromeda.UpdateQuotaUsage,
) ClaimVoucher {
	return &unClaimVoucher{
		voucherRepo:        voucherRepo,
		historyRepo:        historyRepo,
		addVoucherUsage:    addVoucherUsage,
		reduceVoucherUsage: reduceVoucherUsage,
	}
}

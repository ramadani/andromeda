package internal

import "context"

type Voucher struct {
	ID    string `db:"id"`
	Code  string `db:"code"`
	Limit int64  `db:"quota_limit"`
	Usage int64  `db:"quota_usage"`
}

type VoucherRepository interface {
	FindAll(ctx context.Context) ([]*Voucher, error)
	FindByID(ctx context.Context, id string) (*Voucher, error)
	FindByCode(ctx context.Context, code string) (*Voucher, error)
	Update(ctx context.Context, data *Voucher) error
}

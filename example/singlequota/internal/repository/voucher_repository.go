package repository

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/ramadani/andromeda/example/singlequota/internal"
)

type voucherRepositoryMySqlx struct {
	db *sqlx.DB
}

func (r *voucherRepositoryMySqlx) FindAll(ctx context.Context) ([]*internal.Voucher, error) {
	res := make([]*internal.Voucher, 0)
	query := "SELECT * FROM vouchers"
	err := r.db.SelectContext(ctx, &res, query)

	return res, err
}

func (r *voucherRepositoryMySqlx) FindByID(ctx context.Context, id string) (*internal.Voucher, error) {
	res := new(internal.Voucher)
	query := "SELECT * FROM vouchers WHERE id = ?"
	err := r.db.QueryRowxContext(ctx, query, id).StructScan(res)

	return res, err
}

func (r *voucherRepositoryMySqlx) FindByCode(ctx context.Context, code string) (*internal.Voucher, error) {
	res := new(internal.Voucher)
	query := "SELECT * FROM vouchers WHERE code = ?"
	err := r.db.QueryRowxContext(ctx, query, code).StructScan(res)

	return res, err
}

func (r *voucherRepositoryMySqlx) Update(ctx context.Context, data *internal.Voucher) error {
	query := "UPDATE vouchers SET code = ?, quota_limit = ?, quota_usage = ? WHERE id = ?"
	_, err := r.db.ExecContext(ctx, query, data.Code, data.Limit, data.Usage, data.ID)

	return err
}

func NewVoucherRepositoryMySqlx(db *sqlx.DB) internal.VoucherRepository {
	return &voucherRepositoryMySqlx{db: db}
}

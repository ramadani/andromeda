package repository

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/ramadani/andromeda/example/singlequota/internal"
)

type historyRepositoryMySqlx struct {
	db *sqlx.DB
}

func (r *historyRepositoryMySqlx) Create(ctx context.Context, data *internal.History) error {
	query := "INSERT INTO histories (id, voucher_id, user_id, created_at) VALUES (?, ?, ?, ?)"
	_, err := r.db.ExecContext(ctx, query, data.ID, data.VoucherID, data.UserID, data.CreatedAt)

	return err
}

func NewHistoryRepositoryMySqlx(db *sqlx.DB) internal.HistoryRepository {
	return &historyRepositoryMySqlx{db: db}
}

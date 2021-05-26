package internal

import (
	"context"
	"time"
)

type History struct {
	ID        string    `json:"id" db:"id"`
	VoucherID string    `json:"voucherId" db:"voucher_id"`
	UserID    string    `json:"userId" db:"user_id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

type HistoryRepository interface {
	Create(ctx context.Context, data *History) error
}

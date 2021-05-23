package andromeda

import "errors"

var (
	// ErrQuotaNotFound is error for quota not found
	ErrQuotaNotFound = errors.New("quota not found")
	// ErrAddQuotaUsage is error for add quota usage
	ErrAddQuotaUsage = errors.New("error adding quota usage")
	// ErrReduceQuotaUsage is error for reduce quota usage
	ErrReduceQuotaUsage = errors.New("error reducing quota usage")
	// ErrQuotaLimitExceeded is error for quota exceeded
	ErrQuotaLimitExceeded = errors.New("quota limit exceeded")
	// ErrLockedKey .
	ErrLockedKey = errors.New("locked key")
	// ErrMaxRetryExceeded .
	ErrMaxRetryExceeded = errors.New("max retry exceeded")
)

package andromeda_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/ramadani/andromeda"
	"github.com/ramadani/andromeda/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAddQuotaUsage(t *testing.T) {
	ctx := context.TODO()
	mockCtrl := gomock.NewController(t)
	mockCache := mocks.NewMockCache(mockCtrl)
	mockGetQuotaCacheParams := mocks.NewMockGetQuotaCacheParams(mockCtrl)
	mockGetQuotaLimit := mocks.NewMockGetQuota(mockCtrl)
	mockNext := mocks.NewMockUpdateQuotaUsage(mockCtrl)
	option := andromeda.AddUsageOption{Reversible: true}
	addQuotaUsage := andromeda.NewAddQuotaUsage(mockCache, mockGetQuotaCacheParams, mockGetQuotaLimit, mockNext, option)

	t.Run("ErrorGetQuotaCacheParam", func(t *testing.T) {
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		mockErr := errors.New("unexpected")

		mockGetQuotaCacheParams.EXPECT().Do(ctx, quotaReq).Return(nil, mockErr)

		res, err := addQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Nil(t, res)
		assert.EqualError(t, err, mockErr.Error())
	})

	t.Run("DoNextWhenQuotaNotFound", func(t *testing.T) {
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		mockRes := "result"

		mockGetQuotaCacheParams.EXPECT().Do(ctx, quotaReq).Return(nil, andromeda.ErrQuotaNotFound)
		mockNext.EXPECT().Do(ctx, quotaUsageReq).Return(mockRes, nil)

		res, err := addQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Equal(t, mockRes, res)
		assert.Nil(t, err)
	})

	t.Run("ErrorGetQuotaLimit", func(t *testing.T) {
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		mockCacheParams := &andromeda.QuotaCacheParams{
			Key:        "key-123",
			Expiration: 5 * time.Second,
		}
		mockErr := errors.New("unexpected")

		mockGetQuotaCacheParams.EXPECT().Do(ctx, quotaReq).Return(mockCacheParams, nil)
		mockGetQuotaLimit.EXPECT().Do(ctx, quotaReq).Return(int64(0), mockErr)

		res, err := addQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Nil(t, res)
		assert.EqualError(t, err, mockErr.Error())
	})

	t.Run("ErrorIncrementUsage", func(t *testing.T) {
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		mockCacheParams := &andromeda.QuotaCacheParams{
			Key:        "key-123",
			Expiration: 5 * time.Second,
		}
		mockLimit := int64(1000)
		mockErr := errors.New("unexpected")

		mockGetQuotaCacheParams.EXPECT().Do(ctx, quotaReq).Return(mockCacheParams, nil)
		mockGetQuotaLimit.EXPECT().Do(ctx, quotaReq).Return(mockLimit, nil)
		mockCache.EXPECT().IncrBy(ctx, mockCacheParams.Key, quotaUsageReq.Usage).Return(int64(0), mockErr)

		res, err := addQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Nil(t, res)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, andromeda.ErrAddQuotaUsage))
	})

	t.Run("ErrorQuotaLimitExceeded", func(t *testing.T) {
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		mockCacheParams := &andromeda.QuotaCacheParams{
			Key:        "key-123",
			Expiration: 5 * time.Second,
		}
		mockLimit := int64(10000)
		mockUsage := int64(11000)
		mockErr := andromeda.NewQuotaLimitExceededError(mockCacheParams.Key, mockLimit, mockUsage-quotaUsageReq.Usage)

		mockGetQuotaCacheParams.EXPECT().Do(ctx, quotaReq).Return(mockCacheParams, nil)
		mockGetQuotaLimit.EXPECT().Do(ctx, quotaReq).Return(mockLimit, nil)
		mockCache.EXPECT().IncrBy(ctx, mockCacheParams.Key, quotaUsageReq.Usage).Return(mockUsage, nil)
		mockCache.EXPECT().DecrBy(ctx, mockCacheParams.Key, quotaUsageReq.Usage).Return(mockLimit, nil)

		res, err := addQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Nil(t, res)
		assert.EqualError(t, err, mockErr.Error())
	})

	t.Run("ErrorDecrementUsageWhenErrorQuotaLimitExceeded", func(t *testing.T) {
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		mockCacheParams := &andromeda.QuotaCacheParams{
			Key:        "key-123",
			Expiration: 5 * time.Second,
		}
		mockLimit := int64(10000)
		mockUsage := int64(11000)
		mockErr := errors.New("unexpected")

		mockGetQuotaCacheParams.EXPECT().Do(ctx, quotaReq).Return(mockCacheParams, nil)
		mockGetQuotaLimit.EXPECT().Do(ctx, quotaReq).Return(mockLimit, nil)
		mockCache.EXPECT().IncrBy(ctx, mockCacheParams.Key, quotaUsageReq.Usage).Return(mockUsage, nil)
		mockCache.EXPECT().DecrBy(ctx, mockCacheParams.Key, quotaUsageReq.Usage).Return(int64(0), mockErr)

		res, err := addQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Nil(t, res)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, andromeda.ErrReduceQuotaUsage))
	})

	t.Run("DecrementUsageWhenNextHasError", func(t *testing.T) {
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		mockCacheParams := &andromeda.QuotaCacheParams{
			Key:        "key-123",
			Expiration: 5 * time.Second,
		}
		mockLimit := int64(10000)
		mockUsage := int64(10000)
		mockErr := errors.New("unexpected")

		mockGetQuotaCacheParams.EXPECT().Do(ctx, quotaReq).Return(mockCacheParams, nil)
		mockGetQuotaLimit.EXPECT().Do(ctx, quotaReq).Return(mockLimit, nil)
		mockCache.EXPECT().IncrBy(ctx, mockCacheParams.Key, quotaUsageReq.Usage).Return(mockUsage, nil)
		mockNext.EXPECT().Do(ctx, quotaUsageReq).Return(nil, mockErr)
		mockCache.EXPECT().DecrBy(ctx, mockCacheParams.Key, quotaUsageReq.Usage).Return(mockUsage-quotaUsageReq.Usage, nil)

		res, err := addQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Nil(t, res)
		assert.EqualError(t, err, mockErr.Error())
	})

	t.Run("ErrorDecrementUsageWhenNextHasError", func(t *testing.T) {
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		mockCacheParams := &andromeda.QuotaCacheParams{
			Key:        "key-123",
			Expiration: 5 * time.Second,
		}
		mockLimit := int64(10000)
		mockUsage := int64(10000)
		mockErr := errors.New("unexpected")

		mockGetQuotaCacheParams.EXPECT().Do(ctx, quotaReq).Return(mockCacheParams, nil)
		mockGetQuotaLimit.EXPECT().Do(ctx, quotaReq).Return(mockLimit, nil)
		mockCache.EXPECT().IncrBy(ctx, mockCacheParams.Key, quotaUsageReq.Usage).Return(mockUsage, nil)
		mockNext.EXPECT().Do(ctx, quotaUsageReq).Return(nil, mockErr)
		mockCache.EXPECT().DecrBy(ctx, mockCacheParams.Key, quotaUsageReq.Usage).Return(mockUsage-quotaUsageReq.Usage, mockErr)

		res, err := addQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Nil(t, res)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, andromeda.ErrReduceQuotaUsage))
	})

	t.Run("SucceedAddQuotaUsage", func(t *testing.T) {
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		mockCacheParams := &andromeda.QuotaCacheParams{
			Key:        "key-123",
			Expiration: 5 * time.Second,
		}
		mockLimit := int64(10000)
		mockUsage := int64(10000)
		mockRes := "result"

		mockGetQuotaCacheParams.EXPECT().Do(ctx, quotaReq).Return(mockCacheParams, nil)
		mockGetQuotaLimit.EXPECT().Do(ctx, quotaReq).Return(mockLimit, nil)
		mockCache.EXPECT().IncrBy(ctx, mockCacheParams.Key, quotaUsageReq.Usage).Return(mockUsage, nil)
		mockNext.EXPECT().Do(ctx, quotaUsageReq).Return(mockRes, nil)

		res, err := addQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Equal(t, mockRes, res)
		assert.Nil(t, err)
	})

	t.Run("SucceedAddQuotaUsageWithModifiedUsage", func(t *testing.T) {
		opt := andromeda.AddUsageOption{Reversible: true, ModifiedUsage: 1}
		newAddQuotaUsage := andromeda.NewAddQuotaUsage(mockCache, mockGetQuotaCacheParams, mockGetQuotaLimit, mockNext, opt)
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		mockCacheParams := &andromeda.QuotaCacheParams{
			Key:        "key-123",
			Expiration: 5 * time.Second,
		}
		mockLimit := int64(10000)
		mockUsage := int64(10000)
		mockRes := "result"

		mockGetQuotaCacheParams.EXPECT().Do(ctx, quotaReq).Return(mockCacheParams, nil)
		mockGetQuotaLimit.EXPECT().Do(ctx, quotaReq).Return(mockLimit, nil)
		mockCache.EXPECT().IncrBy(ctx, mockCacheParams.Key, opt.ModifiedUsage).Return(mockUsage, nil)
		mockNext.EXPECT().Do(ctx, quotaUsageReq).Return(mockRes, nil)

		res, err := newAddQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Equal(t, mockRes, res)
		assert.Nil(t, err)
	})

	t.Run("ListenOnError", func(t *testing.T) {
		mockListener := mocks.NewMockUpdateQuotaUsageListener(mockCtrl)
		opt := andromeda.AddUsageOption{Reversible: true, Listener: mockListener}
		newAddQuotaUsage := andromeda.NewAddQuotaUsage(mockCache, mockGetQuotaCacheParams, mockGetQuotaLimit, mockNext, opt)
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		mockCacheParams := &andromeda.QuotaCacheParams{
			Key:        "key-123",
			Expiration: 5 * time.Second,
		}
		mockLimit := int64(10000)
		mockUsage := int64(11000)
		mockErr := errors.New("unexpected")
		onErr := fmt.Errorf("%w: %v", andromeda.ErrReduceQuotaUsage, mockErr)

		mockGetQuotaCacheParams.EXPECT().Do(ctx, quotaReq).Return(mockCacheParams, nil)
		mockGetQuotaLimit.EXPECT().Do(ctx, quotaReq).Return(mockLimit, nil)
		mockCache.EXPECT().IncrBy(ctx, mockCacheParams.Key, quotaUsageReq.Usage).Return(mockUsage, nil)
		mockCache.EXPECT().DecrBy(ctx, mockCacheParams.Key, quotaUsageReq.Usage).Return(int64(0), mockErr)
		mockListener.EXPECT().OnError(ctx, quotaUsageReq, onErr)

		res, err := newAddQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Nil(t, res)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, andromeda.ErrReduceQuotaUsage))
	})

	t.Run("NoListenOnNextError", func(t *testing.T) {
		mockListener := mocks.NewMockUpdateQuotaUsageListener(mockCtrl)
		opt := andromeda.AddUsageOption{Reversible: true, Listener: mockListener}
		newAddQuotaUsage := andromeda.NewAddQuotaUsage(mockCache, mockGetQuotaCacheParams, mockGetQuotaLimit, mockNext, opt)
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		mockCacheParams := &andromeda.QuotaCacheParams{
			Key:        "key-123",
			Expiration: 5 * time.Second,
		}
		mockLimit := int64(10000)
		mockUsage := int64(10000)
		mockErr := errors.New("unexpected")

		mockGetQuotaCacheParams.EXPECT().Do(ctx, quotaReq).Return(mockCacheParams, nil)
		mockGetQuotaLimit.EXPECT().Do(ctx, quotaReq).Return(mockLimit, nil)
		mockCache.EXPECT().IncrBy(ctx, mockCacheParams.Key, quotaUsageReq.Usage).Return(mockUsage, nil)
		mockNext.EXPECT().Do(ctx, quotaUsageReq).Return(nil, mockErr)
		mockCache.EXPECT().DecrBy(ctx, mockCacheParams.Key, quotaUsageReq.Usage).Return(mockUsage-quotaUsageReq.Usage, nil)

		res, err := newAddQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Nil(t, res)
		assert.EqualError(t, err, mockErr.Error())
	})

	t.Run("ListenOnSuccess", func(t *testing.T) {
		mockListener := mocks.NewMockUpdateQuotaUsageListener(mockCtrl)
		opt := andromeda.AddUsageOption{Reversible: true, Listener: mockListener}
		newAddQuotaUsage := andromeda.NewAddQuotaUsage(mockCache, mockGetQuotaCacheParams, mockGetQuotaLimit, mockNext, opt)
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		mockCacheParams := &andromeda.QuotaCacheParams{
			Key:        "key-123",
			Expiration: 5 * time.Second,
		}
		mockLimit := int64(10000)
		mockUsage := int64(10000)
		mockRes := "result"

		mockGetQuotaCacheParams.EXPECT().Do(ctx, quotaReq).Return(mockCacheParams, nil)
		mockGetQuotaLimit.EXPECT().Do(ctx, quotaReq).Return(mockLimit, nil)
		mockCache.EXPECT().IncrBy(ctx, mockCacheParams.Key, quotaUsageReq.Usage).Return(mockUsage, nil)
		mockNext.EXPECT().Do(ctx, quotaUsageReq).Return(mockRes, nil)
		mockListener.EXPECT().OnSuccess(ctx, quotaUsageReq, mockUsage)

		res, err := newAddQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Equal(t, mockRes, res)
		assert.Nil(t, err)
	})
}

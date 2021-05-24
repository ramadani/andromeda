package andromeda_test

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/ramadani/andromeda"
	"github.com/ramadani/andromeda/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestReduceQuotaUsage(t *testing.T) {
	ctx := context.TODO()
	mockCtrl := gomock.NewController(t)
	mockCache := mocks.NewMockCache(mockCtrl)
	mockGetQuotaCacheParams := mocks.NewMockGetQuotaCacheParams(mockCtrl)
	mockNext := mocks.NewMockUpdateQuotaUsage(mockCtrl)
	option := andromeda.ReduceUsageOption{Reversible: true}
	reduceQuotaUsage := andromeda.NewReduceQuotaUsage(mockCache, mockGetQuotaCacheParams, mockNext, option)

	t.Run("ErrorGetQuotaCacheParam", func(t *testing.T) {
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		mockErr := errors.New("unexpected")

		mockGetQuotaCacheParams.EXPECT().Do(ctx, quotaReq).Return(nil, mockErr)

		res, err := reduceQuotaUsage.Do(ctx, quotaUsageReq)

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

		res, err := reduceQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Equal(t, mockRes, res)
		assert.Nil(t, err)
	})

	t.Run("ErrorDecrementUsage", func(t *testing.T) {
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		mockCacheParams := &andromeda.QuotaCacheParams{
			Key:        "key-123",
			Expiration: 5 * time.Second,
		}
		mockErr := errors.New("unexpected")

		mockGetQuotaCacheParams.EXPECT().Do(ctx, quotaReq).Return(mockCacheParams, nil)
		mockCache.EXPECT().DecrBy(ctx, mockCacheParams.Key, quotaUsageReq.Usage).Return(int64(0), mockErr)

		res, err := reduceQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Nil(t, res)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, andromeda.ErrReduceQuotaUsage))
	})

	t.Run("IncrementUsageWhenNextHasError", func(t *testing.T) {
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		mockCacheParams := &andromeda.QuotaCacheParams{
			Key:        "key-123",
			Expiration: 5 * time.Second,
		}
		mockErr := errors.New("unexpected")

		mockGetQuotaCacheParams.EXPECT().Do(ctx, quotaReq).Return(mockCacheParams, nil)
		mockCache.EXPECT().DecrBy(ctx, mockCacheParams.Key, quotaUsageReq.Usage).Return(int64(0), nil)
		mockNext.EXPECT().Do(ctx, quotaUsageReq).Return(nil, mockErr)
		mockCache.EXPECT().IncrBy(ctx, mockCacheParams.Key, quotaUsageReq.Usage).Return(quotaUsageReq.Usage, nil)

		res, err := reduceQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Nil(t, res)
		assert.EqualError(t, err, mockErr.Error())
	})

	t.Run("ErrorIncrementUsageWhenNextHasError", func(t *testing.T) {
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		mockCacheParams := &andromeda.QuotaCacheParams{
			Key:        "key-123",
			Expiration: 5 * time.Second,
		}
		mockErr := errors.New("unexpected")

		mockGetQuotaCacheParams.EXPECT().Do(ctx, quotaReq).Return(mockCacheParams, nil)
		mockCache.EXPECT().DecrBy(ctx, mockCacheParams.Key, quotaUsageReq.Usage).Return(int64(0), nil)
		mockNext.EXPECT().Do(ctx, quotaUsageReq).Return(nil, mockErr)
		mockCache.EXPECT().IncrBy(ctx, mockCacheParams.Key, quotaUsageReq.Usage).Return(int64(0), mockErr)

		res, err := reduceQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Nil(t, res)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, andromeda.ErrAddQuotaUsage))
	})

	t.Run("SucceedReduceQuotaUsage", func(t *testing.T) {
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		mockCacheParams := &andromeda.QuotaCacheParams{
			Key:        "key-123",
			Expiration: 5 * time.Second,
		}
		mockRes := "result"

		mockGetQuotaCacheParams.EXPECT().Do(ctx, quotaReq).Return(mockCacheParams, nil)
		mockCache.EXPECT().DecrBy(ctx, mockCacheParams.Key, quotaUsageReq.Usage).Return(int64(0), nil)
		mockNext.EXPECT().Do(ctx, quotaUsageReq).Return(mockRes, nil)

		res, err := reduceQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Equal(t, mockRes, res)
		assert.Nil(t, err)
	})

	t.Run("SucceedReduceQuotaUsageWithModifiedUsage", func(t *testing.T) {
		opt := andromeda.ReduceUsageOption{Reversible: true, ModifiedUsage: 1}
		newReduceQuotaUsage := andromeda.NewReduceQuotaUsage(mockCache, mockGetQuotaCacheParams, mockNext, opt)
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		mockCacheParams := &andromeda.QuotaCacheParams{
			Key:        "key-123",
			Expiration: 5 * time.Second,
		}
		mockRes := "result"

		mockGetQuotaCacheParams.EXPECT().Do(ctx, quotaReq).Return(mockCacheParams, nil)
		mockCache.EXPECT().DecrBy(ctx, mockCacheParams.Key, opt.ModifiedUsage).Return(int64(0), nil)
		mockNext.EXPECT().Do(ctx, quotaUsageReq).Return(mockRes, nil)

		res, err := newReduceQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Equal(t, mockRes, res)
		assert.Nil(t, err)
	})
}

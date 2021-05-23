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

		id := "123"
		value := int64(1000)
		mockErr := errors.New("unexpected")

		mockGetQuotaCacheParams.EXPECT().Do(ctx, id, nil).Return(nil, mockErr)

		res, err := addQuotaUsage.Do(ctx, id, value, nil)

		assert.Nil(t, res)
		assert.EqualError(t, err, mockErr.Error())
	})

	t.Run("DoNextWhenQuotaNotFound", func(t *testing.T) {
		defer mockCtrl.Finish()

		id := "123"
		value := int64(1000)
		mockRes := "result"

		mockGetQuotaCacheParams.EXPECT().Do(ctx, id, nil).Return(nil, andromeda.ErrQuotaNotFound)
		mockNext.EXPECT().Do(ctx, id, value, nil).Return(mockRes, nil)

		res, err := addQuotaUsage.Do(ctx, id, value, nil)

		assert.Equal(t, mockRes, res)
		assert.Nil(t, err)
	})

	t.Run("ErrorGetQuotaLimit", func(t *testing.T) {
		defer mockCtrl.Finish()

		id := "123"
		value := int64(1000)
		mockCacheParams := &andromeda.QuotaCacheParams{
			Key:        "key-123",
			Expiration: 5 * time.Second,
		}
		mockErr := errors.New("unexpected")

		mockGetQuotaCacheParams.EXPECT().Do(ctx, id, nil).Return(mockCacheParams, nil)
		mockGetQuotaLimit.EXPECT().Do(ctx, id, nil).Return(int64(0), mockErr)

		res, err := addQuotaUsage.Do(ctx, id, value, nil)

		assert.Nil(t, res)
		assert.EqualError(t, err, mockErr.Error())
	})

	t.Run("ErrorIncrementUsage", func(t *testing.T) {
		defer mockCtrl.Finish()

		id := "123"
		value := int64(1000)
		mockCacheParams := &andromeda.QuotaCacheParams{
			Key:        "key-123",
			Expiration: 5 * time.Second,
		}
		mockLimit := int64(1000)
		mockErr := errors.New("unexpected")

		mockGetQuotaCacheParams.EXPECT().Do(ctx, id, nil).Return(mockCacheParams, nil)
		mockGetQuotaLimit.EXPECT().Do(ctx, id, nil).Return(mockLimit, nil)
		mockCache.EXPECT().IncrBy(ctx, mockCacheParams.Key, value).Return(int64(0), mockErr)

		res, err := addQuotaUsage.Do(ctx, id, value, nil)

		assert.Nil(t, res)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, andromeda.ErrAddQuotaUsage))
	})

	t.Run("ErrorQuotaLimitExceeded", func(t *testing.T) {
		defer mockCtrl.Finish()

		id := "123"
		value := int64(1000)
		mockCacheParams := &andromeda.QuotaCacheParams{
			Key:        "key-123",
			Expiration: 5 * time.Second,
		}
		mockLimit := int64(10000)
		mockUsage := int64(11000)
		mockErr := andromeda.NewQuotaLimitExceededError(mockCacheParams.Key, mockLimit, mockUsage-value)

		mockGetQuotaCacheParams.EXPECT().Do(ctx, id, nil).Return(mockCacheParams, nil)
		mockGetQuotaLimit.EXPECT().Do(ctx, id, nil).Return(mockLimit, nil)
		mockCache.EXPECT().IncrBy(ctx, mockCacheParams.Key, value).Return(mockUsage, nil)
		mockCache.EXPECT().DecrBy(ctx, mockCacheParams.Key, value).Return(mockLimit, nil)

		res, err := addQuotaUsage.Do(ctx, id, value, nil)

		assert.Nil(t, res)
		assert.EqualError(t, err, mockErr.Error())
	})

	t.Run("ErrorDecrementUsageWhenErrorQuotaLimitExceeded", func(t *testing.T) {
		defer mockCtrl.Finish()

		id := "123"
		value := int64(1000)
		mockCacheParams := &andromeda.QuotaCacheParams{
			Key:        "key-123",
			Expiration: 5 * time.Second,
		}
		mockLimit := int64(10000)
		mockUsage := int64(11000)
		mockErr := errors.New("unexpected")

		mockGetQuotaCacheParams.EXPECT().Do(ctx, id, nil).Return(mockCacheParams, nil)
		mockGetQuotaLimit.EXPECT().Do(ctx, id, nil).Return(mockLimit, nil)
		mockCache.EXPECT().IncrBy(ctx, mockCacheParams.Key, value).Return(mockUsage, nil)
		mockCache.EXPECT().DecrBy(ctx, mockCacheParams.Key, value).Return(int64(0), mockErr)

		res, err := addQuotaUsage.Do(ctx, id, value, nil)

		assert.Nil(t, res)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, andromeda.ErrReduceQuotaUsage))
	})

	t.Run("DecrementUsageWhenNextHasError", func(t *testing.T) {
		defer mockCtrl.Finish()

		id := "123"
		value := int64(1000)
		mockCacheParams := &andromeda.QuotaCacheParams{
			Key:        "key-123",
			Expiration: 5 * time.Second,
		}
		mockLimit := int64(10000)
		mockUsage := int64(10000)
		mockErr := errors.New("unexpected")

		mockGetQuotaCacheParams.EXPECT().Do(ctx, id, nil).Return(mockCacheParams, nil)
		mockGetQuotaLimit.EXPECT().Do(ctx, id, nil).Return(mockLimit, nil)
		mockCache.EXPECT().IncrBy(ctx, mockCacheParams.Key, value).Return(mockUsage, nil)
		mockNext.EXPECT().Do(ctx, id, value, nil).Return(nil, mockErr)
		mockCache.EXPECT().DecrBy(ctx, mockCacheParams.Key, value).Return(mockUsage-value, nil)

		res, err := addQuotaUsage.Do(ctx, id, value, nil)

		assert.Nil(t, res)
		assert.EqualError(t, err, mockErr.Error())
	})

	t.Run("ErrorDecrementUsageWhenNextHasError", func(t *testing.T) {
		defer mockCtrl.Finish()

		id := "123"
		value := int64(1000)
		mockCacheParams := &andromeda.QuotaCacheParams{
			Key:        "key-123",
			Expiration: 5 * time.Second,
		}
		mockLimit := int64(10000)
		mockUsage := int64(10000)
		mockErr := errors.New("unexpected")

		mockGetQuotaCacheParams.EXPECT().Do(ctx, id, nil).Return(mockCacheParams, nil)
		mockGetQuotaLimit.EXPECT().Do(ctx, id, nil).Return(mockLimit, nil)
		mockCache.EXPECT().IncrBy(ctx, mockCacheParams.Key, value).Return(mockUsage, nil)
		mockNext.EXPECT().Do(ctx, id, value, nil).Return(nil, mockErr)
		mockCache.EXPECT().DecrBy(ctx, mockCacheParams.Key, value).Return(mockUsage-value, mockErr)

		res, err := addQuotaUsage.Do(ctx, id, value, nil)

		assert.Nil(t, res)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, andromeda.ErrReduceQuotaUsage))
	})

	t.Run("SucceedAddQuotaUsage", func(t *testing.T) {
		defer mockCtrl.Finish()

		id := "123"
		value := int64(1000)
		mockCacheParams := &andromeda.QuotaCacheParams{
			Key:        "key-123",
			Expiration: 5 * time.Second,
		}
		mockLimit := int64(10000)
		mockUsage := int64(10000)
		mockRes := "result"

		mockGetQuotaCacheParams.EXPECT().Do(ctx, id, nil).Return(mockCacheParams, nil)
		mockGetQuotaLimit.EXPECT().Do(ctx, id, nil).Return(mockLimit, nil)
		mockCache.EXPECT().IncrBy(ctx, mockCacheParams.Key, value).Return(mockUsage, nil)
		mockNext.EXPECT().Do(ctx, id, value, nil).Return(mockRes, nil)

		res, err := addQuotaUsage.Do(ctx, id, value, nil)

		assert.Equal(t, mockRes, res)
		assert.Nil(t, err)
	})

	t.Run("SucceedAddQuotaUsageWithModifiedUsage", func(t *testing.T) {
		opt := andromeda.AddUsageOption{Reversible: true, ModifiedUsage: 1}
		addQuotaUsage = andromeda.NewAddQuotaUsage(mockCache, mockGetQuotaCacheParams, mockGetQuotaLimit, mockNext, opt)
		defer mockCtrl.Finish()

		id := "123"
		value := int64(1000)
		mockCacheParams := &andromeda.QuotaCacheParams{
			Key:        "key-123",
			Expiration: 5 * time.Second,
		}
		mockLimit := int64(10000)
		mockUsage := int64(10000)
		mockRes := "result"

		mockGetQuotaCacheParams.EXPECT().Do(ctx, id, nil).Return(mockCacheParams, nil)
		mockGetQuotaLimit.EXPECT().Do(ctx, id, nil).Return(mockLimit, nil)
		mockCache.EXPECT().IncrBy(ctx, mockCacheParams.Key, opt.ModifiedUsage).Return(mockUsage, nil)
		mockNext.EXPECT().Do(ctx, id, value, nil).Return(mockRes, nil)

		res, err := addQuotaUsage.Do(ctx, id, value, nil)

		assert.Equal(t, mockRes, res)
		assert.Nil(t, err)
	})
}

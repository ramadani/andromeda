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

func TestGetCachedQuota(t *testing.T) {
	ctx := context.TODO()
	mockCtrl := gomock.NewController(t)
	mockCache := mocks.NewMockCache(mockCtrl)
	mockGetQuotaCacheParams := mocks.NewMockGetQuotaCacheParams(mockCtrl)
	getCachedQuota := andromeda.NewGetCachedQuota(mockCache, mockGetQuotaCacheParams)

	t.Run("ErrorGetQuotaCacheParams", func(t *testing.T) {
		defer mockCtrl.Finish()

		req := &andromeda.QuotaRequest{QuotaID: "123"}
		mockErr := errors.New("unexpected")

		mockGetQuotaCacheParams.EXPECT().Do(ctx, req).Return(nil, mockErr)

		res, err := getCachedQuota.Do(ctx, req)

		assert.Equal(t, int64(0), res)
		assert.EqualError(t, err, mockErr.Error())
	})

	t.Run("ErrorGetCache", func(t *testing.T) {
		defer mockCtrl.Finish()

		req := &andromeda.QuotaRequest{QuotaID: "123"}
		mockCacheParams := &andromeda.QuotaCacheParams{Key: "123-key", Expiration: 5 * time.Minute}
		mockErr := errors.New("unexpected")

		mockGetQuotaCacheParams.EXPECT().Do(ctx, req).Return(mockCacheParams, nil)
		mockCache.EXPECT().Get(ctx, mockCacheParams.Key).Return("", mockErr)

		res, err := getCachedQuota.Do(ctx, req)

		assert.Equal(t, int64(0), res)
		assert.EqualError(t, err, mockErr.Error())
	})

	t.Run("ErrorQuotaNotFound", func(t *testing.T) {
		defer mockCtrl.Finish()

		req := &andromeda.QuotaRequest{QuotaID: "123"}
		mockCacheParams := &andromeda.QuotaCacheParams{Key: "123-key", Expiration: 5 * time.Minute}

		mockGetQuotaCacheParams.EXPECT().Do(ctx, req).Return(mockCacheParams, nil)
		mockCache.EXPECT().Get(ctx, mockCacheParams.Key).Return("", andromeda.ErrCacheNotFound)

		res, err := getCachedQuota.Do(ctx, req)

		assert.Equal(t, int64(0), res)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, andromeda.ErrQuotaNotFound))
	})

	t.Run("SucceedGetQuota", func(t *testing.T) {
		defer mockCtrl.Finish()

		req := &andromeda.QuotaRequest{QuotaID: "123"}
		mockCacheParams := &andromeda.QuotaCacheParams{Key: "123-key", Expiration: 5 * time.Minute}

		mockGetQuotaCacheParams.EXPECT().Do(ctx, req).Return(mockCacheParams, nil)
		mockCache.EXPECT().Get(ctx, mockCacheParams.Key).Return("1000", nil)

		res, err := getCachedQuota.Do(ctx, req)

		assert.Equal(t, int64(1000), res)
		assert.Nil(t, err)
	})
}

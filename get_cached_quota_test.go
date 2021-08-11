package andromeda_test

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/ramadani/andromeda"
	"github.com/ramadani/andromeda/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetCachedQuota(t *testing.T) {
	ctx := context.TODO()
	mockCtrl := gomock.NewController(t)
	mockCache := mocks.NewMockCache(mockCtrl)
	mockGetQuotaKey := mocks.NewMockGetQuotaKey(mockCtrl)
	getCachedQuota := andromeda.NewGetCachedQuota(mockCache, mockGetQuotaKey)

	t.Run("ErrorGetQuotaKey", func(t *testing.T) {
		defer mockCtrl.Finish()

		req := &andromeda.QuotaRequest{QuotaID: "123"}
		mockErr := errors.New("unexpected")

		mockGetQuotaKey.EXPECT().Do(ctx, req).Return("", mockErr)

		res, err := getCachedQuota.Do(ctx, req)

		assert.Equal(t, int64(0), res)
		assert.EqualError(t, err, mockErr.Error())
	})

	t.Run("ErrorGetCache", func(t *testing.T) {
		defer mockCtrl.Finish()

		req := &andromeda.QuotaRequest{QuotaID: "123"}
		key := "123-key"
		mockErr := errors.New("unexpected")

		mockGetQuotaKey.EXPECT().Do(ctx, req).Return(key, nil)
		mockCache.EXPECT().Get(ctx, key).Return("", mockErr)

		res, err := getCachedQuota.Do(ctx, req)

		assert.Equal(t, int64(0), res)
		assert.EqualError(t, err, mockErr.Error())
	})

	t.Run("ErrorQuotaNotFound", func(t *testing.T) {
		defer mockCtrl.Finish()

		req := &andromeda.QuotaRequest{QuotaID: "123"}
		key := "123-key"

		mockGetQuotaKey.EXPECT().Do(ctx, req).Return(key, nil)
		mockCache.EXPECT().Get(ctx, key).Return("", andromeda.ErrCacheNotFound)

		res, err := getCachedQuota.Do(ctx, req)

		assert.Equal(t, int64(0), res)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, andromeda.ErrQuotaNotFound))
	})

	t.Run("SucceedGetQuota", func(t *testing.T) {
		defer mockCtrl.Finish()

		req := &andromeda.QuotaRequest{QuotaID: "123"}
		key := "123-key"

		mockGetQuotaKey.EXPECT().Do(ctx, req).Return(key, nil)
		mockCache.EXPECT().Get(ctx, key).Return("1000", nil)

		res, err := getCachedQuota.Do(ctx, req)

		assert.Equal(t, int64(1000), res)
		assert.Nil(t, err)
	})

	t.Run("ErrorConvertValue", func(t *testing.T) {
		defer mockCtrl.Finish()

		req := &andromeda.QuotaRequest{QuotaID: "123"}
		key := "123-key"

		mockGetQuotaKey.EXPECT().Do(ctx, req).Return(key, nil)
		mockCache.EXPECT().Get(ctx, key).Return("lorem", nil)

		res, err := getCachedQuota.Do(ctx, req)

		assert.Equal(t, int64(0), res)
		assert.NotNil(t, err)
	})
}

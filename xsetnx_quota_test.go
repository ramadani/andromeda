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

func TestXSetNXQuota(t *testing.T) {
	ctx := context.TODO()
	mockCtrl := gomock.NewController(t)
	mockCache := mocks.NewMockCache(mockCtrl)
	mockGetQuotaCacheParams := mocks.NewMockGetQuotaCacheParams(mockCtrl)
	mockGetQuota := mocks.NewMockGetQuota(mockCtrl)
	lockIn := time.Second * 5
	xSetNXQuota := andromeda.NewXSetNXQuota(mockCache, mockGetQuotaCacheParams, mockGetQuota, lockIn)

	t.Run("ErrorGetQuotaCacheParams", func(t *testing.T) {
		defer mockCtrl.Finish()

		req := &andromeda.QuotaRequest{QuotaID: "123"}
		mockErr := errors.New("unexpected")

		mockGetQuotaCacheParams.EXPECT().Do(ctx, req).Return(nil, mockErr)

		err := xSetNXQuota.Do(ctx, req)

		assert.EqualError(t, err, mockErr.Error())
	})

	t.Run("ErrorCacheExists", func(t *testing.T) {
		defer mockCtrl.Finish()

		req := &andromeda.QuotaRequest{QuotaID: "123"}
		mockCacheParams := &andromeda.QuotaCacheParams{Key: "123-key", Expiration: 5 * time.Minute}
		mockErr := errors.New("unexpected")

		mockGetQuotaCacheParams.EXPECT().Do(ctx, req).Return(mockCacheParams, nil)
		mockCache.EXPECT().Exists(ctx, mockCacheParams.Key).Return(int64(0), mockErr)

		err := xSetNXQuota.Do(ctx, req)

		assert.EqualError(t, err, mockErr.Error())
	})

	t.Run("NoErrorWhenCacheExists", func(t *testing.T) {
		defer mockCtrl.Finish()

		req := &andromeda.QuotaRequest{QuotaID: "123"}
		mockCacheParams := &andromeda.QuotaCacheParams{Key: "123-key", Expiration: 5 * time.Minute}

		mockGetQuotaCacheParams.EXPECT().Do(ctx, req).Return(mockCacheParams, nil)
		mockCache.EXPECT().Exists(ctx, mockCacheParams.Key).Return(int64(1), nil)

		err := xSetNXQuota.Do(ctx, req)

		assert.Nil(t, err)
	})

	t.Run("ErrorLockKey", func(t *testing.T) {
		defer mockCtrl.Finish()

		req := &andromeda.QuotaRequest{QuotaID: "123"}
		mockCacheParams := &andromeda.QuotaCacheParams{Key: "123-key", Expiration: 5 * time.Minute}
		mockLockKey := fmt.Sprintf("%s-lock", mockCacheParams.Key)
		mockErr := errors.New("unexpected")

		mockGetQuotaCacheParams.EXPECT().Do(ctx, req).Return(mockCacheParams, nil)
		mockCache.EXPECT().Exists(ctx, mockCacheParams.Key).Return(int64(0), nil)
		mockCache.EXPECT().SetNX(ctx, mockLockKey, 1, lockIn).Return(false, mockErr)

		err := xSetNXQuota.Do(ctx, req)

		assert.EqualError(t, err, mockErr.Error())
	})

	t.Run("ErrLockedKey", func(t *testing.T) {
		defer mockCtrl.Finish()

		req := &andromeda.QuotaRequest{QuotaID: "123"}
		mockCacheParams := &andromeda.QuotaCacheParams{Key: "123-key", Expiration: 5 * time.Minute}
		mockLockKey := fmt.Sprintf("%s-lock", mockCacheParams.Key)

		mockGetQuotaCacheParams.EXPECT().Do(ctx, req).Return(mockCacheParams, nil)
		mockCache.EXPECT().Exists(ctx, mockCacheParams.Key).Return(int64(0), nil)
		mockCache.EXPECT().SetNX(ctx, mockLockKey, 1, lockIn).Return(false, nil)

		err := xSetNXQuota.Do(ctx, req)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, andromeda.ErrLockedKey))
	})

	t.Run("ErrorNextDoAndUnlock", func(t *testing.T) {
		defer mockCtrl.Finish()

		req := &andromeda.QuotaRequest{QuotaID: "123"}
		mockCacheParams := &andromeda.QuotaCacheParams{Key: "123-key", Expiration: 5 * time.Minute}
		mockLockKey := fmt.Sprintf("%s-lock", mockCacheParams.Key)
		mockErr := errors.New("unexpected")

		mockGetQuotaCacheParams.EXPECT().Do(ctx, req).Return(mockCacheParams, nil)
		mockCache.EXPECT().Exists(ctx, mockCacheParams.Key).Return(int64(0), nil)
		mockCache.EXPECT().SetNX(ctx, mockLockKey, 1, lockIn).Return(true, nil)
		mockGetQuota.EXPECT().Do(ctx, req).Return(int64(0), mockErr)
		mockCache.EXPECT().Del(ctx, mockLockKey).Return(int64(1), nil)

		err := xSetNXQuota.Do(ctx, req)

		assert.EqualError(t, err, mockErr.Error())
	})

	t.Run("ErrorNextDoAndErrorUnlock", func(t *testing.T) {
		defer mockCtrl.Finish()

		req := &andromeda.QuotaRequest{QuotaID: "123"}
		mockCacheParams := &andromeda.QuotaCacheParams{Key: "123-key", Expiration: 5 * time.Minute}
		mockLockKey := fmt.Sprintf("%s-lock", mockCacheParams.Key)
		mockErr := errors.New("unexpected")
		mockErrUnlock := errors.New("err unlock")

		mockGetQuotaCacheParams.EXPECT().Do(ctx, req).Return(mockCacheParams, nil)
		mockCache.EXPECT().Exists(ctx, mockCacheParams.Key).Return(int64(0), nil)
		mockCache.EXPECT().SetNX(ctx, mockLockKey, 1, lockIn).Return(true, nil)
		mockGetQuota.EXPECT().Do(ctx, req).Return(int64(0), mockErr)
		mockCache.EXPECT().Del(ctx, mockLockKey).Return(int64(0), mockErrUnlock)

		err := xSetNXQuota.Do(ctx, req)

		assert.EqualError(t, err, mockErrUnlock.Error())
	})

	t.Run("ErrorSetQuotaAndUnlock", func(t *testing.T) {
		defer mockCtrl.Finish()

		req := &andromeda.QuotaRequest{QuotaID: "123"}
		mockCacheParams := &andromeda.QuotaCacheParams{Key: "123-key", Expiration: 5 * time.Minute}
		mockLockKey := fmt.Sprintf("%s-lock", mockCacheParams.Key)
		mockErr := errors.New("unexpected")

		mockGetQuotaCacheParams.EXPECT().Do(ctx, req).Return(mockCacheParams, nil)
		mockCache.EXPECT().Exists(ctx, mockCacheParams.Key).Return(int64(0), nil)
		mockCache.EXPECT().SetNX(ctx, mockLockKey, 1, lockIn).Return(true, nil)
		mockGetQuota.EXPECT().Do(ctx, req).Return(int64(1000), nil)
		mockCache.EXPECT().SetNX(ctx, mockCacheParams.Key, int64(1000), mockCacheParams.Expiration).Return(false, mockErr)
		mockCache.EXPECT().Del(ctx, mockLockKey).Return(int64(1), nil)

		err := xSetNXQuota.Do(ctx, req)

		assert.EqualError(t, err, mockErr.Error())
	})

	t.Run("ErrorSetQuotaAndErrorUnlock", func(t *testing.T) {
		defer mockCtrl.Finish()

		req := &andromeda.QuotaRequest{QuotaID: "123"}
		mockCacheParams := &andromeda.QuotaCacheParams{Key: "123-key", Expiration: 5 * time.Minute}
		mockLockKey := fmt.Sprintf("%s-lock", mockCacheParams.Key)
		mockErr := errors.New("unexpected")
		mockErrUnlock := errors.New("err unlock")

		mockGetQuotaCacheParams.EXPECT().Do(ctx, req).Return(mockCacheParams, nil)
		mockCache.EXPECT().Exists(ctx, mockCacheParams.Key).Return(int64(0), nil)
		mockCache.EXPECT().SetNX(ctx, mockLockKey, 1, lockIn).Return(true, nil)
		mockGetQuota.EXPECT().Do(ctx, req).Return(int64(1000), nil)
		mockCache.EXPECT().SetNX(ctx, mockCacheParams.Key, int64(1000), mockCacheParams.Expiration).Return(false, mockErr)
		mockCache.EXPECT().Del(ctx, mockLockKey).Return(int64(0), mockErrUnlock)

		err := xSetNXQuota.Do(ctx, req)

		assert.EqualError(t, err, mockErrUnlock.Error())
	})

	t.Run("SucceedSetQuota", func(t *testing.T) {
		defer mockCtrl.Finish()

		req := &andromeda.QuotaRequest{QuotaID: "123"}
		mockCacheParams := &andromeda.QuotaCacheParams{Key: "123-key", Expiration: 5 * time.Minute}
		mockLockKey := fmt.Sprintf("%s-lock", mockCacheParams.Key)

		mockGetQuotaCacheParams.EXPECT().Do(ctx, req).Return(mockCacheParams, nil)
		mockCache.EXPECT().Exists(ctx, mockCacheParams.Key).Return(int64(0), nil)
		mockCache.EXPECT().SetNX(ctx, mockLockKey, 1, lockIn).Return(true, nil)
		mockGetQuota.EXPECT().Do(ctx, req).Return(int64(1000), nil)
		mockCache.EXPECT().SetNX(ctx, mockCacheParams.Key, int64(1000), mockCacheParams.Expiration).Return(false, nil)
		mockCache.EXPECT().Del(ctx, mockLockKey).Return(int64(0), nil)

		err := xSetNXQuota.Do(ctx, req)

		assert.Nil(t, err)
	})
}

func TestRetryableXSetNXQuota(t *testing.T) {
	ctx := context.TODO()
	mockCtrl := gomock.NewController(t)
	mockNext := mocks.NewMockXSetNXQuota(mockCtrl)
	maxRetry := 5
	sleepIn := time.Duration(0)
	retryable := andromeda.NewRetryableXSetNXQuota(mockNext, maxRetry, sleepIn)

	t.Run("ErrMaxRetryExceeded", func(t *testing.T) {
		req := &andromeda.QuotaRequest{QuotaID: "123"}
		mockErr := andromeda.ErrMaxRetryExceeded
		mockErrNext := errors.New("unexpected")

		mockNext.EXPECT().Do(ctx, req).Return(mockErrNext).Times(maxRetry)

		err := retryable.Do(ctx, req)

		assert.EqualError(t, err, mockErr.Error())
	})

	t.Run("PartialError", func(t *testing.T) {
		req := &andromeda.QuotaRequest{QuotaID: "123"}
		mockErrNext := errors.New("unexpected")

		mockNext.EXPECT().Do(ctx, req).Return(mockErrNext).Times(maxRetry - 2)
		mockNext.EXPECT().Do(ctx, req).Return(nil)

		err := retryable.Do(ctx, req)

		assert.Nil(t, err)
	})

	t.Run("NoError", func(t *testing.T) {
		req := &andromeda.QuotaRequest{QuotaID: "123"}
		mockNext.EXPECT().Do(ctx, req).Return(nil)

		err := retryable.Do(ctx, req)

		assert.Nil(t, err)
	})
}

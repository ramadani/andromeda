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
	mockGetQuotaKey := mocks.NewMockGetQuotaKey(mockCtrl)
	mockGetQuotaExp := mocks.NewMockGetQuotaExpiration(mockCtrl)
	mockGetQuota := mocks.NewMockGetQuota(mockCtrl)
	lockIn := time.Second * 5
	xSetNXQuota := andromeda.NewXSetNXQuota(mockCache, mockGetQuotaKey, mockGetQuotaExp, mockGetQuota, lockIn)

	t.Run("ErrorGetQuotaKey", func(t *testing.T) {
		defer mockCtrl.Finish()

		req := &andromeda.QuotaRequest{QuotaID: "123"}
		mockErr := errors.New("unexpected")

		mockGetQuotaKey.EXPECT().Do(ctx, req).Return("", mockErr)

		err := xSetNXQuota.Do(ctx, req)

		assert.EqualError(t, err, mockErr.Error())
	})

	t.Run("IgnoreWhenQuotaKeyNotFound", func(t *testing.T) {
		defer mockCtrl.Finish()

		req := &andromeda.QuotaRequest{QuotaID: "123"}
		mockErr := andromeda.ErrQuotaNotFound

		mockGetQuotaKey.EXPECT().Do(ctx, req).Return("", mockErr)

		err := xSetNXQuota.Do(ctx, req)

		assert.Nil(t, err)
	})

	t.Run("ErrorCacheExists", func(t *testing.T) {
		defer mockCtrl.Finish()

		req := &andromeda.QuotaRequest{QuotaID: "123"}
		key := "123-key"
		mockErr := errors.New("unexpected")

		mockGetQuotaKey.EXPECT().Do(ctx, req).Return(key, nil)
		mockCache.EXPECT().Exists(ctx, key).Return(int64(0), mockErr)

		err := xSetNXQuota.Do(ctx, req)

		assert.EqualError(t, err, mockErr.Error())
	})

	t.Run("NoErrorWhenCacheExists", func(t *testing.T) {
		defer mockCtrl.Finish()

		req := &andromeda.QuotaRequest{QuotaID: "123"}
		key := "123-key"

		mockGetQuotaKey.EXPECT().Do(ctx, req).Return(key, nil)
		mockCache.EXPECT().Exists(ctx, key).Return(int64(1), nil)

		err := xSetNXQuota.Do(ctx, req)

		assert.Nil(t, err)
	})

	t.Run("ErrorLockKey", func(t *testing.T) {
		defer mockCtrl.Finish()

		req := &andromeda.QuotaRequest{QuotaID: "123"}
		key := "123-key"
		mockLockKey := fmt.Sprintf("%s-lock", key)
		mockErr := errors.New("unexpected")

		mockGetQuotaKey.EXPECT().Do(ctx, req).Return(key, nil)
		mockCache.EXPECT().Exists(ctx, key).Return(int64(0), nil)
		mockCache.EXPECT().SetNX(ctx, mockLockKey, 1, lockIn).Return(false, mockErr)

		err := xSetNXQuota.Do(ctx, req)

		assert.EqualError(t, err, mockErr.Error())
	})

	t.Run("ErrLockedKey", func(t *testing.T) {
		defer mockCtrl.Finish()

		req := &andromeda.QuotaRequest{QuotaID: "123"}
		key := "123-key"
		mockLockKey := fmt.Sprintf("%s-lock", key)

		mockGetQuotaKey.EXPECT().Do(ctx, req).Return(key, nil)
		mockCache.EXPECT().Exists(ctx, key).Return(int64(0), nil)
		mockCache.EXPECT().SetNX(ctx, mockLockKey, 1, lockIn).Return(false, nil)

		err := xSetNXQuota.Do(ctx, req)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, andromeda.ErrLockedKey))
	})

	t.Run("ErrorGetQuotaAndUnlock", func(t *testing.T) {
		defer mockCtrl.Finish()

		req := &andromeda.QuotaRequest{QuotaID: "123"}
		key := "123-key"
		mockLockKey := fmt.Sprintf("%s-lock", key)
		mockErr := errors.New("unexpected")

		mockGetQuotaKey.EXPECT().Do(ctx, req).Return(key, nil)
		mockCache.EXPECT().Exists(ctx, key).Return(int64(0), nil)
		mockCache.EXPECT().SetNX(ctx, mockLockKey, 1, lockIn).Return(true, nil)
		mockGetQuota.EXPECT().Do(ctx, req).Return(int64(0), mockErr)
		mockCache.EXPECT().Del(ctx, mockLockKey).Return(int64(1), nil)

		err := xSetNXQuota.Do(ctx, req)

		assert.EqualError(t, err, mockErr.Error())
	})

	t.Run("ErrorGetQuotaAndErrorUnlock", func(t *testing.T) {
		defer mockCtrl.Finish()

		req := &andromeda.QuotaRequest{QuotaID: "123"}
		key := "123-key"
		mockLockKey := fmt.Sprintf("%s-lock", key)
		mockErr := errors.New("unexpected")
		mockErrUnlock := errors.New("err unlock")

		mockGetQuotaKey.EXPECT().Do(ctx, req).Return(key, nil)
		mockCache.EXPECT().Exists(ctx, key).Return(int64(0), nil)
		mockCache.EXPECT().SetNX(ctx, mockLockKey, 1, lockIn).Return(true, nil)
		mockGetQuota.EXPECT().Do(ctx, req).Return(int64(0), mockErr)
		mockCache.EXPECT().Del(ctx, mockLockKey).Return(int64(0), mockErrUnlock)

		err := xSetNXQuota.Do(ctx, req)

		assert.EqualError(t, err, mockErrUnlock.Error())
	})

	t.Run("ErrorGetQuotaExpirationAndUnlock", func(t *testing.T) {
		defer mockCtrl.Finish()

		req := &andromeda.QuotaRequest{QuotaID: "123"}
		key := "123-key"
		mockLockKey := fmt.Sprintf("%s-lock", key)
		mockErr := errors.New("unexpected")

		mockGetQuotaKey.EXPECT().Do(ctx, req).Return(key, nil)
		mockCache.EXPECT().Exists(ctx, key).Return(int64(0), nil)
		mockCache.EXPECT().SetNX(ctx, mockLockKey, 1, lockIn).Return(true, nil)
		mockGetQuota.EXPECT().Do(ctx, req).Return(int64(1000), nil)
		mockGetQuotaExp.EXPECT().Do(ctx, req).Return(time.Duration(0), mockErr)
		mockCache.EXPECT().Del(ctx, mockLockKey).Return(int64(1), nil)

		err := xSetNXQuota.Do(ctx, req)

		assert.EqualError(t, err, mockErr.Error())
	})

	t.Run("ErrorSetQuotaAndUnlock", func(t *testing.T) {
		defer mockCtrl.Finish()

		req := &andromeda.QuotaRequest{QuotaID: "123"}
		key := "123-key"
		exp := time.Hour * 1
		mockLockKey := fmt.Sprintf("%s-lock", key)
		mockErr := errors.New("unexpected")

		mockGetQuotaKey.EXPECT().Do(ctx, req).Return(key, nil)
		mockCache.EXPECT().Exists(ctx, key).Return(int64(0), nil)
		mockCache.EXPECT().SetNX(ctx, mockLockKey, 1, lockIn).Return(true, nil)
		mockGetQuota.EXPECT().Do(ctx, req).Return(int64(1000), nil)
		mockGetQuotaExp.EXPECT().Do(ctx, req).Return(exp, nil)
		mockCache.EXPECT().SetNX(ctx, key, int64(1000), exp).Return(false, mockErr)
		mockCache.EXPECT().Del(ctx, mockLockKey).Return(int64(1), nil)

		err := xSetNXQuota.Do(ctx, req)

		assert.EqualError(t, err, mockErr.Error())
	})

	t.Run("ErrorSetQuotaAndErrorUnlock", func(t *testing.T) {
		defer mockCtrl.Finish()

		req := &andromeda.QuotaRequest{QuotaID: "123"}
		key := "123-key"
		exp := time.Hour * 1
		mockLockKey := fmt.Sprintf("%s-lock", key)
		mockErr := errors.New("unexpected")
		mockErrUnlock := errors.New("err unlock")

		mockGetQuotaKey.EXPECT().Do(ctx, req).Return(key, nil)
		mockCache.EXPECT().Exists(ctx, key).Return(int64(0), nil)
		mockCache.EXPECT().SetNX(ctx, mockLockKey, 1, lockIn).Return(true, nil)
		mockGetQuota.EXPECT().Do(ctx, req).Return(int64(1000), nil)
		mockGetQuotaExp.EXPECT().Do(ctx, req).Return(exp, nil)
		mockCache.EXPECT().SetNX(ctx, key, int64(1000), exp).Return(false, mockErr)
		mockCache.EXPECT().Del(ctx, mockLockKey).Return(int64(0), mockErrUnlock)

		err := xSetNXQuota.Do(ctx, req)

		assert.EqualError(t, err, mockErrUnlock.Error())
	})

	t.Run("SucceedSetQuota", func(t *testing.T) {
		defer mockCtrl.Finish()

		req := &andromeda.QuotaRequest{QuotaID: "123"}
		key := "123-key"
		exp := time.Hour * 1
		mockLockKey := fmt.Sprintf("%s-lock", key)

		mockGetQuotaKey.EXPECT().Do(ctx, req).Return(key, nil)
		mockCache.EXPECT().Exists(ctx, key).Return(int64(0), nil)
		mockCache.EXPECT().SetNX(ctx, mockLockKey, 1, lockIn).Return(true, nil)
		mockGetQuota.EXPECT().Do(ctx, req).Return(int64(1000), nil)
		mockGetQuotaExp.EXPECT().Do(ctx, req).Return(exp, nil)
		mockCache.EXPECT().SetNX(ctx, key, int64(1000), exp).Return(false, nil)
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
	retryIn := time.Duration(0)
	retryable := andromeda.NewRetryableXSetNXQuota(mockNext, maxRetry, retryIn)

	t.Run("ErrMaxRetryExceeded", func(t *testing.T) {
		req := &andromeda.QuotaRequest{QuotaID: "123"}
		mockErr := andromeda.ErrMaxRetryExceeded
		mockErrNext := errors.New("unexpected")

		mockNext.EXPECT().Do(ctx, req).Return(mockErrNext).Times(maxRetry)

		err := retryable.Do(ctx, req)

		assert.True(t, errors.Is(err, mockErr))
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

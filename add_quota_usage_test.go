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
)

func TestAddQuotaUsage(t *testing.T) {
	ctx := context.TODO()
	mockCtrl := gomock.NewController(t)
	mockCache := mocks.NewMockCache(mockCtrl)
	mockGetQuotaUsageKey := mocks.NewMockGetQuotaKey(mockCtrl)
	mockGetQuotaLimit := mocks.NewMockGetQuota(mockCtrl)
	mockNext := mocks.NewMockUpdateQuotaUsage(mockCtrl)
	option := andromeda.AddUsageOption{}
	addQuotaUsage := andromeda.NewAddQuotaUsage(mockCache, mockGetQuotaUsageKey, mockGetQuotaLimit, mockNext, option)

	t.Run("ErrorGetQuotaUsageKey", func(t *testing.T) {
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		mockErr := errors.New("unexpected")

		mockGetQuotaUsageKey.EXPECT().Do(ctx, quotaReq).Return("", mockErr)

		res, err := addQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Nil(t, res)
		assert.EqualError(t, err, mockErr.Error())
	})

	t.Run("DoNextWhenQuotaNotFound", func(t *testing.T) {
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		mockRes := "result"

		mockGetQuotaUsageKey.EXPECT().Do(ctx, quotaReq).Return("", andromeda.ErrQuotaNotFound)
		mockNext.EXPECT().Do(ctx, quotaUsageReq).Return(mockRes, nil)

		res, err := addQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Equal(t, mockRes, res)
		assert.Nil(t, err)
	})

	t.Run("DoNextWhenQuotaNotFoundWithFormat", func(t *testing.T) {
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		mockRes := "result"

		mockGetQuotaUsageKey.EXPECT().Do(ctx, quotaReq).Return("", fmt.Errorf("error: %w", andromeda.ErrQuotaNotFound))
		mockNext.EXPECT().Do(ctx, quotaUsageReq).Return(mockRes, nil)

		res, err := addQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Equal(t, mockRes, res)
		assert.Nil(t, err)
	})

	t.Run("ErrorGetQuotaLimit", func(t *testing.T) {
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		key := "key-123"
		mockErr := errors.New("unexpected")

		mockGetQuotaUsageKey.EXPECT().Do(ctx, quotaReq).Return(key, nil)
		mockGetQuotaLimit.EXPECT().Do(ctx, quotaReq).Return(int64(0), mockErr)

		res, err := addQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Nil(t, res)
		assert.EqualError(t, err, mockErr.Error())
	})

	t.Run("ErrorIncrementUsage", func(t *testing.T) {
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		key := "key-123"
		mockLimit := int64(1000)
		mockErr := errors.New("unexpected")

		mockGetQuotaUsageKey.EXPECT().Do(ctx, quotaReq).Return(key, nil)
		mockGetQuotaLimit.EXPECT().Do(ctx, quotaReq).Return(mockLimit, nil)
		mockCache.EXPECT().IncrBy(ctx, key, quotaUsageReq.Usage).Return(int64(0), mockErr)

		res, err := addQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Nil(t, res)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, andromeda.ErrAddQuotaUsage))
	})

	t.Run("ErrorQuotaLimitExceeded", func(t *testing.T) {
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		key := "key-123"
		mockLimit := int64(10000)
		mockUsage := int64(11000)
		mockErr := andromeda.NewQuotaLimitExceededError(key, mockLimit, mockUsage-quotaUsageReq.Usage)

		mockGetQuotaUsageKey.EXPECT().Do(ctx, quotaReq).Return(key, nil)
		mockGetQuotaLimit.EXPECT().Do(ctx, quotaReq).Return(mockLimit, nil)
		mockCache.EXPECT().IncrBy(ctx, key, quotaUsageReq.Usage).Return(mockUsage, nil)
		mockCache.EXPECT().DecrBy(ctx, key, quotaUsageReq.Usage).Return(mockLimit, nil)

		res, err := addQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Nil(t, res)
		assert.EqualError(t, err, mockErr.Error())
	})

	t.Run("ErrorDecrementUsageWhenErrorQuotaLimitExceeded", func(t *testing.T) {
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		key := "key-123"
		mockLimit := int64(10000)
		mockUsage := int64(11000)
		mockErr := errors.New("unexpected")

		mockGetQuotaUsageKey.EXPECT().Do(ctx, quotaReq).Return(key, nil)
		mockGetQuotaLimit.EXPECT().Do(ctx, quotaReq).Return(mockLimit, nil)
		mockCache.EXPECT().IncrBy(ctx, key, quotaUsageReq.Usage).Return(mockUsage, nil)
		mockCache.EXPECT().DecrBy(ctx, key, quotaUsageReq.Usage).Return(int64(0), mockErr)

		res, err := addQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Nil(t, res)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, andromeda.ErrReduceQuotaUsage))
	})

	t.Run("DecrementUsageWhenNextHasError", func(t *testing.T) {
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		key := "key-123"
		mockLimit := int64(10000)
		mockUsage := int64(10000)
		mockErr := errors.New("unexpected")

		mockGetQuotaUsageKey.EXPECT().Do(ctx, quotaReq).Return(key, nil)
		mockGetQuotaLimit.EXPECT().Do(ctx, quotaReq).Return(mockLimit, nil)
		mockCache.EXPECT().IncrBy(ctx, key, quotaUsageReq.Usage).Return(mockUsage, nil)
		mockNext.EXPECT().Do(ctx, quotaUsageReq).Return(nil, mockErr)
		mockCache.EXPECT().DecrBy(ctx, key, quotaUsageReq.Usage).Return(mockUsage-quotaUsageReq.Usage, nil)

		res, err := addQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Nil(t, res)
		assert.EqualError(t, err, mockErr.Error())
	})

	t.Run("ErrorDecrementUsageWhenNextHasError", func(t *testing.T) {
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		key := "key-123"
		mockLimit := int64(10000)
		mockUsage := int64(10000)
		mockErr := errors.New("unexpected")

		mockGetQuotaUsageKey.EXPECT().Do(ctx, quotaReq).Return(key, nil)
		mockGetQuotaLimit.EXPECT().Do(ctx, quotaReq).Return(mockLimit, nil)
		mockCache.EXPECT().IncrBy(ctx, key, quotaUsageReq.Usage).Return(mockUsage, nil)
		mockNext.EXPECT().Do(ctx, quotaUsageReq).Return(nil, mockErr)
		mockCache.EXPECT().DecrBy(ctx, key, quotaUsageReq.Usage).Return(mockUsage-quotaUsageReq.Usage, mockErr)

		res, err := addQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Nil(t, res)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, andromeda.ErrReduceQuotaUsage))
	})

	t.Run("SucceedAddQuotaUsage", func(t *testing.T) {
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		key := "key-123"
		mockLimit := int64(10000)
		mockUsage := int64(10000)
		mockRes := "result"

		mockGetQuotaUsageKey.EXPECT().Do(ctx, quotaReq).Return(key, nil)
		mockGetQuotaLimit.EXPECT().Do(ctx, quotaReq).Return(mockLimit, nil)
		mockCache.EXPECT().IncrBy(ctx, key, quotaUsageReq.Usage).Return(mockUsage, nil)
		mockNext.EXPECT().Do(ctx, quotaUsageReq).Return(mockRes, nil)

		res, err := addQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Equal(t, mockRes, res)
		assert.Nil(t, err)
	})

	t.Run("SucceedAddQuotaUsageWithModifiedUsage", func(t *testing.T) {
		opt := andromeda.AddUsageOption{ModifiedUsage: 1}
		newAddQuotaUsage := andromeda.NewAddQuotaUsage(mockCache, mockGetQuotaUsageKey, mockGetQuotaLimit, mockNext, opt)
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		key := "key-123"
		mockLimit := int64(10000)
		mockUsage := int64(10000)
		mockRes := "result"

		mockGetQuotaUsageKey.EXPECT().Do(ctx, quotaReq).Return(key, nil)
		mockGetQuotaLimit.EXPECT().Do(ctx, quotaReq).Return(mockLimit, nil)
		mockCache.EXPECT().IncrBy(ctx, key, opt.ModifiedUsage).Return(mockUsage, nil)
		mockNext.EXPECT().Do(ctx, quotaUsageReq).Return(mockRes, nil)

		res, err := newAddQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Equal(t, mockRes, res)
		assert.Nil(t, err)
	})

	t.Run("ListenOnError", func(t *testing.T) {
		mockListener := mocks.NewMockUpdateQuotaUsageListener(mockCtrl)
		opt := andromeda.AddUsageOption{Listener: mockListener}
		newAddQuotaUsage := andromeda.NewAddQuotaUsage(mockCache, mockGetQuotaUsageKey, mockGetQuotaLimit, mockNext, opt)
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		key := "key-123"
		mockLimit := int64(10000)
		mockUsage := int64(11000)
		mockErr := errors.New("unexpected")
		onErr := fmt.Errorf("%w: %v", andromeda.ErrReduceQuotaUsage, mockErr)

		mockGetQuotaUsageKey.EXPECT().Do(ctx, quotaReq).Return(key, nil)
		mockGetQuotaLimit.EXPECT().Do(ctx, quotaReq).Return(mockLimit, nil)
		mockCache.EXPECT().IncrBy(ctx, key, quotaUsageReq.Usage).Return(mockUsage, nil)
		mockCache.EXPECT().DecrBy(ctx, key, quotaUsageReq.Usage).Return(int64(0), mockErr)
		mockListener.EXPECT().OnError(ctx, quotaUsageReq, onErr)

		res, err := newAddQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Nil(t, res)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, andromeda.ErrReduceQuotaUsage))
	})

	t.Run("NoListenOnNextError", func(t *testing.T) {
		mockListener := mocks.NewMockUpdateQuotaUsageListener(mockCtrl)
		opt := andromeda.AddUsageOption{Listener: mockListener}
		newAddQuotaUsage := andromeda.NewAddQuotaUsage(mockCache, mockGetQuotaUsageKey, mockGetQuotaLimit, mockNext, opt)
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		key := "key-123"
		mockLimit := int64(10000)
		mockUsage := int64(10000)
		mockErr := errors.New("unexpected")

		mockGetQuotaUsageKey.EXPECT().Do(ctx, quotaReq).Return(key, nil)
		mockGetQuotaLimit.EXPECT().Do(ctx, quotaReq).Return(mockLimit, nil)
		mockCache.EXPECT().IncrBy(ctx, key, quotaUsageReq.Usage).Return(mockUsage, nil)
		mockNext.EXPECT().Do(ctx, quotaUsageReq).Return(nil, mockErr)
		mockCache.EXPECT().DecrBy(ctx, key, quotaUsageReq.Usage).Return(mockUsage-quotaUsageReq.Usage, nil)

		res, err := newAddQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Nil(t, res)
		assert.EqualError(t, err, mockErr.Error())
	})

	t.Run("ListenOnSuccess", func(t *testing.T) {
		mockListener := mocks.NewMockUpdateQuotaUsageListener(mockCtrl)
		opt := andromeda.AddUsageOption{Listener: mockListener}
		newAddQuotaUsage := andromeda.NewAddQuotaUsage(mockCache, mockGetQuotaUsageKey, mockGetQuotaLimit, mockNext, opt)
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		key := "key-123"
		mockLimit := int64(10000)
		mockUsage := int64(10000)
		mockRes := "result"

		mockGetQuotaUsageKey.EXPECT().Do(ctx, quotaReq).Return(key, nil)
		mockGetQuotaLimit.EXPECT().Do(ctx, quotaReq).Return(mockLimit, nil)
		mockCache.EXPECT().IncrBy(ctx, key, quotaUsageReq.Usage).Return(mockUsage, nil)
		mockNext.EXPECT().Do(ctx, quotaUsageReq).Return(mockRes, nil)
		mockListener.EXPECT().OnSuccess(ctx, quotaUsageReq, mockUsage)

		res, err := newAddQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Equal(t, mockRes, res)
		assert.Nil(t, err)
	})
}

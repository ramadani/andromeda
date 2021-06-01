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

func TestReduceQuotaUsage(t *testing.T) {
	ctx := context.TODO()
	mockCtrl := gomock.NewController(t)
	mockCache := mocks.NewMockCache(mockCtrl)
	mockGetQuotaUsageKey := mocks.NewMockGetQuotaKey(mockCtrl)
	mockNext := mocks.NewMockUpdateQuotaUsage(mockCtrl)
	option := andromeda.ReduceUsageOption{}
	reduceQuotaUsage := andromeda.NewReduceQuotaUsage(mockCache, mockGetQuotaUsageKey, mockNext, option)

	t.Run("ErrorGetQuotaUsageKey", func(t *testing.T) {
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		mockErr := errors.New("unexpected")

		mockGetQuotaUsageKey.EXPECT().Do(ctx, quotaReq).Return("", mockErr)

		res, err := reduceQuotaUsage.Do(ctx, quotaUsageReq)

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

		res, err := reduceQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Equal(t, mockRes, res)
		assert.Nil(t, err)
	})

	t.Run("ErrorDecrementUsage", func(t *testing.T) {
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		key := "key-123"
		mockErr := errors.New("unexpected")

		mockGetQuotaUsageKey.EXPECT().Do(ctx, quotaReq).Return(key, nil)
		mockCache.EXPECT().DecrBy(ctx, key, quotaUsageReq.Usage).Return(int64(0), mockErr)

		res, err := reduceQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Nil(t, res)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, andromeda.ErrReduceQuotaUsage))
	})

	t.Run("ReverseQuotaUsageWhenTotalUsageLessThanZero", func(t *testing.T) {
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		key := "key-123"

		mockGetQuotaUsageKey.EXPECT().Do(ctx, quotaReq).Return(key, nil)
		mockCache.EXPECT().DecrBy(ctx, key, quotaUsageReq.Usage).Return(int64(-1), nil)
		mockCache.EXPECT().IncrBy(ctx, key, quotaUsageReq.Usage).Return(int64(999), nil)

		res, err := reduceQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Nil(t, res)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, andromeda.ErrInvalidMinQuotaUsage))
	})

	t.Run("ErrorReverseQuotaUsageWhenTotalUsageLessThanZero", func(t *testing.T) {
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		key := "key-123"
		mockErr := errors.New("unexpected")

		mockGetQuotaUsageKey.EXPECT().Do(ctx, quotaReq).Return(key, nil)
		mockCache.EXPECT().DecrBy(ctx, key, quotaUsageReq.Usage).Return(int64(-1), nil)
		mockCache.EXPECT().IncrBy(ctx, key, quotaUsageReq.Usage).Return(int64(0), mockErr)

		res, err := reduceQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Nil(t, res)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, andromeda.ErrAddQuotaUsage))
	})

	t.Run("IncrementUsageWhenNextHasError", func(t *testing.T) {
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		key := "key-123"
		mockErr := errors.New("unexpected")

		mockGetQuotaUsageKey.EXPECT().Do(ctx, quotaReq).Return(key, nil)
		mockCache.EXPECT().DecrBy(ctx, key, quotaUsageReq.Usage).Return(int64(0), nil)
		mockNext.EXPECT().Do(ctx, quotaUsageReq).Return(nil, mockErr)
		mockCache.EXPECT().IncrBy(ctx, key, quotaUsageReq.Usage).Return(quotaUsageReq.Usage, nil)

		res, err := reduceQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Nil(t, res)
		assert.EqualError(t, err, mockErr.Error())
	})

	t.Run("ErrorIncrementUsageWhenNextHasError", func(t *testing.T) {
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		key := "key-123"
		mockErr := errors.New("unexpected")

		mockGetQuotaUsageKey.EXPECT().Do(ctx, quotaReq).Return(key, nil)
		mockCache.EXPECT().DecrBy(ctx, key, quotaUsageReq.Usage).Return(int64(0), nil)
		mockNext.EXPECT().Do(ctx, quotaUsageReq).Return(nil, mockErr)
		mockCache.EXPECT().IncrBy(ctx, key, quotaUsageReq.Usage).Return(int64(0), mockErr)

		res, err := reduceQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Nil(t, res)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, andromeda.ErrAddQuotaUsage))
	})

	t.Run("SucceedReduceQuotaUsage", func(t *testing.T) {
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		key := "key-123"
		mockRes := "result"

		mockGetQuotaUsageKey.EXPECT().Do(ctx, quotaReq).Return(key, nil)
		mockCache.EXPECT().DecrBy(ctx, key, quotaUsageReq.Usage).Return(int64(0), nil)
		mockNext.EXPECT().Do(ctx, quotaUsageReq).Return(mockRes, nil)

		res, err := reduceQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Equal(t, mockRes, res)
		assert.Nil(t, err)
	})

	t.Run("SucceedReduceQuotaUsageWithModifiedUsage", func(t *testing.T) {
		opt := andromeda.ReduceUsageOption{ModifiedUsage: 1}
		newReduceQuotaUsage := andromeda.NewReduceQuotaUsage(mockCache, mockGetQuotaUsageKey, mockNext, opt)
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		key := "key-123"
		mockRes := "result"

		mockGetQuotaUsageKey.EXPECT().Do(ctx, quotaReq).Return(key, nil)
		mockCache.EXPECT().DecrBy(ctx, key, opt.ModifiedUsage).Return(int64(0), nil)
		mockNext.EXPECT().Do(ctx, quotaUsageReq).Return(mockRes, nil)

		res, err := newReduceQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Equal(t, mockRes, res)
		assert.Nil(t, err)
	})

	t.Run("ListenOnError", func(t *testing.T) {
		mockListener := mocks.NewMockUpdateQuotaUsageListener(mockCtrl)
		opt := andromeda.ReduceUsageOption{Listener: mockListener}
		newReduceQuotaUsage := andromeda.NewReduceQuotaUsage(mockCache, mockGetQuotaUsageKey, mockNext, opt)
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		key := "key-123"
		mockErr := errors.New("unexpected")
		theErr := fmt.Errorf("%w: %v", andromeda.ErrReduceQuotaUsage, mockErr)

		mockGetQuotaUsageKey.EXPECT().Do(ctx, quotaReq).Return(key, nil)
		mockCache.EXPECT().DecrBy(ctx, key, quotaUsageReq.Usage).Return(int64(0), mockErr)
		mockListener.EXPECT().OnError(ctx, quotaUsageReq, theErr)

		res, err := newReduceQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Nil(t, res)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, andromeda.ErrReduceQuotaUsage))
	})

	t.Run("NoListenOnNextError", func(t *testing.T) {
		mockListener := mocks.NewMockUpdateQuotaUsageListener(mockCtrl)
		opt := andromeda.ReduceUsageOption{Listener: mockListener}
		newReduceQuotaUsage := andromeda.NewReduceQuotaUsage(mockCache, mockGetQuotaUsageKey, mockNext, opt)
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		key := "key-123"
		mockErr := errors.New("unexpected")

		mockGetQuotaUsageKey.EXPECT().Do(ctx, quotaReq).Return(key, nil)
		mockCache.EXPECT().DecrBy(ctx, key, quotaUsageReq.Usage).Return(int64(0), nil)
		mockNext.EXPECT().Do(ctx, quotaUsageReq).Return(nil, mockErr)
		mockCache.EXPECT().IncrBy(ctx, key, quotaUsageReq.Usage).Return(quotaUsageReq.Usage, nil)

		res, err := newReduceQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Nil(t, res)
		assert.EqualError(t, err, mockErr.Error())
	})

	t.Run("ListenOnSuccess", func(t *testing.T) {
		mockListener := mocks.NewMockUpdateQuotaUsageListener(mockCtrl)
		opt := andromeda.ReduceUsageOption{Listener: mockListener}
		newReduceQuotaUsage := andromeda.NewReduceQuotaUsage(mockCache, mockGetQuotaUsageKey, mockNext, opt)
		defer mockCtrl.Finish()

		quotaUsageReq := &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: int64(1000)}
		quotaReq := &andromeda.QuotaRequest{QuotaID: quotaUsageReq.QuotaID, Data: quotaUsageReq.Data}
		key := "key-123"
		mockRes := "result"

		mockGetQuotaUsageKey.EXPECT().Do(ctx, quotaReq).Return(key, nil)
		mockCache.EXPECT().DecrBy(ctx, key, quotaUsageReq.Usage).Return(int64(10), nil)
		mockNext.EXPECT().Do(ctx, quotaUsageReq).Return(mockRes, nil)
		mockListener.EXPECT().OnSuccess(ctx, quotaUsageReq, int64(10))

		res, err := newReduceQuotaUsage.Do(ctx, quotaUsageReq)

		assert.Equal(t, mockRes, res)
		assert.Nil(t, err)
	})
}

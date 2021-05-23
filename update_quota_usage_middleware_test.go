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

func TestUpdateQuotaUsageMiddleware(t *testing.T) {
	ctx := context.TODO()
	mockCtrl := gomock.NewController(t)
	mockPrev := mocks.NewMockUpdateQuotaUsage(mockCtrl)
	mockNext := mocks.NewMockUpdateQuotaUsage(mockCtrl)
	updateQuotaUsage := andromeda.NewUpdateQuotaUsageMiddleware(mockPrev, mockNext)

	t.Run("ExitWhenPrevHasError", func(t *testing.T) {
		id := "123"
		value := int64(1000)
		mockErr := errors.New("unexpected")

		mockPrev.EXPECT().Do(ctx, id, value, nil).Return(nil, mockErr)

		res, err := updateQuotaUsage.Do(ctx, id, value, nil)

		assert.Nil(t, res)
		assert.EqualError(t, err, mockErr.Error())
	})

	t.Run("NextHasError", func(t *testing.T) {
		id := "123"
		value := int64(1000)
		mockRes := "result"
		mockErr := errors.New("unexpected")

		mockPrev.EXPECT().Do(ctx, id, value, nil).Return(mockRes, nil)
		mockNext.EXPECT().Do(ctx, id, value, nil).Return(nil, mockErr)

		res, err := updateQuotaUsage.Do(ctx, id, value, nil)

		assert.Nil(t, res)
		assert.EqualError(t, err, mockErr.Error())
	})

	t.Run("NoError", func(t *testing.T) {
		id := "123"
		value := int64(1000)
		mockRes := "result"

		mockPrev.EXPECT().Do(ctx, id, value, nil).Return(mockRes, nil)
		mockNext.EXPECT().Do(ctx, id, value, nil).Return(mockRes, nil)

		res, err := updateQuotaUsage.Do(ctx, id, value, nil)

		assert.Equal(t, mockRes, res)
		assert.Nil(t, err)
	})
}

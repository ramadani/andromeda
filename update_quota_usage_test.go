package andromeda_test

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/ramadani/andromeda"
	"github.com/ramadani/andromeda/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNopUpdateQuotaUsage(t *testing.T) {
	ctx := context.TODO()
	nopUpdateQuotaUsage := andromeda.NopUpdateQuotaUsage()

	t.Run("DoNothing", func(t *testing.T) {
		res, err := nopUpdateQuotaUsage.Do(ctx, new(andromeda.QuotaUsageRequest))

		assert.Nil(t, res)
		assert.Nil(t, err)
	})
}

func TestXSetNXQuotaUsage(t *testing.T) {
	ctx := context.TODO()
	mockCtrl := gomock.NewController(t)
	mockXSetNXQuota := mocks.NewMockXSetNXQuota(mockCtrl)
	xSetNXQuotaUsage := andromeda.NewXSetNXQuotaUsage(mockXSetNXQuota)

	t.Run("Do", func(t *testing.T) {
		mockXSetNXQuota.EXPECT().Do(ctx, new(andromeda.QuotaRequest)).Return(nil)

		res, err := xSetNXQuotaUsage.Do(ctx, new(andromeda.QuotaUsageRequest))

		assert.Nil(t, res)
		assert.Nil(t, err)
	})
}

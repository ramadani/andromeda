package andromeda_test

import (
	"github.com/golang/mock/gomock"
	"github.com/ramadani/andromeda"
	"github.com/ramadani/andromeda/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAndromedaAddQuotaUsage(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockCache := mocks.NewMockCache(mockCtrl)
	mockGetQuotaLimit := mocks.NewMockGetQuota(mockCtrl)
	mockGetQuotaCacheParams := mocks.NewMockGetQuotaCacheParams(mockCtrl)

	t.Run("ConfigWithoutNextUpdateQuotaUsage", func(t *testing.T) {
		defer mockCtrl.Finish()

		conf := andromeda.AddQuotaUsageConfig{
			Cache:               mockCache,
			GetQuotaLimit:       mockGetQuotaLimit,
			GetQuotaCacheParams: mockGetQuotaCacheParams,
		}

		addQuotaUsage := andromeda.AddQuotaUsage(conf)

		_, ok := addQuotaUsage.(andromeda.UpdateQuotaUsage)

		assert.True(t, ok)
	})

	t.Run("ConfigWithNextUpdateQuotaUsage", func(t *testing.T) {
		mockNext := mocks.NewMockUpdateQuotaUsage(mockCtrl)

		defer mockCtrl.Finish()

		conf := andromeda.AddQuotaUsageConfig{
			Cache:               mockCache,
			GetQuotaLimit:       mockGetQuotaLimit,
			GetQuotaCacheParams: mockGetQuotaCacheParams,
			Next:                mockNext,
		}

		addQuotaUsage := andromeda.AddQuotaUsage(conf)

		_, ok := addQuotaUsage.(andromeda.UpdateQuotaUsage)

		assert.True(t, ok)
	})

	t.Run("ConfigWithGetQuotaUsage", func(t *testing.T) {
		mockNext := mocks.NewMockUpdateQuotaUsage(mockCtrl)
		mockGetQuotaUsage := mocks.NewMockGetQuota(mockCtrl)

		defer mockCtrl.Finish()

		conf := andromeda.AddQuotaUsageConfig{
			Cache:               mockCache,
			GetQuotaLimit:       mockGetQuotaLimit,
			GetQuotaUsage:       mockGetQuotaUsage,
			LockInGetQuotaUsage: 5 * time.Second,
			GetQuotaCacheParams: mockGetQuotaCacheParams,
			Next:                mockNext,
		}

		addQuotaUsage := andromeda.AddQuotaUsage(conf)

		_, ok := addQuotaUsage.(andromeda.UpdateQuotaUsage)

		assert.True(t, ok)
	})
}

func TestAndromedaReduceQuotaUsage(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockCache := mocks.NewMockCache(mockCtrl)
	mockGetQuotaCacheParams := mocks.NewMockGetQuotaCacheParams(mockCtrl)

	t.Run("ConfigWithoutNextUpdateQuotaUsage", func(t *testing.T) {
		defer mockCtrl.Finish()

		conf := andromeda.ReduceQuotaUsageConfig{
			Cache:               mockCache,
			GetQuotaCacheParams: mockGetQuotaCacheParams,
		}

		reduceQuotaUsage := andromeda.ReduceQuotaUsage(conf)

		_, ok := reduceQuotaUsage.(andromeda.UpdateQuotaUsage)

		assert.True(t, ok)
	})

	t.Run("ConfigWithNextUpdateQuotaUsage", func(t *testing.T) {
		mockNext := mocks.NewMockUpdateQuotaUsage(mockCtrl)

		defer mockCtrl.Finish()

		conf := andromeda.ReduceQuotaUsageConfig{
			Cache:               mockCache,
			GetQuotaCacheParams: mockGetQuotaCacheParams,
			Next:                mockNext,
		}

		reduceQuotaUsage := andromeda.ReduceQuotaUsage(conf)

		_, ok := reduceQuotaUsage.(andromeda.UpdateQuotaUsage)

		assert.True(t, ok)
	})

	t.Run("ConfigWithGetQuotaUsage", func(t *testing.T) {
		mockNext := mocks.NewMockUpdateQuotaUsage(mockCtrl)
		mockGetQuotaUsage := mocks.NewMockGetQuota(mockCtrl)

		defer mockCtrl.Finish()

		conf := andromeda.ReduceQuotaUsageConfig{
			Cache:               mockCache,
			GetQuotaUsage:       mockGetQuotaUsage,
			LockInGetQuotaUsage: 5 * time.Second,
			GetQuotaCacheParams: mockGetQuotaCacheParams,
			Next:                mockNext,
		}

		reduceQuotaUsage := andromeda.ReduceQuotaUsage(conf)

		_, ok := reduceQuotaUsage.(andromeda.UpdateQuotaUsage)

		assert.True(t, ok)
	})
}

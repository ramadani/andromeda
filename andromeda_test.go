package andromeda_test

import (
	"context"
	"fmt"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/golang/mock/gomock"
	"github.com/ramadani/andromeda"
	"github.com/ramadani/andromeda/cache"
	"github.com/ramadani/andromeda/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGetQuotaUsageConfig(t *testing.T) {
	tests := []struct {
		name     string
		conf     andromeda.GetQuotaUsageConfig
		lockIn   time.Duration
		maxRetry int
		retryIn  time.Duration
	}{
		{
			name:     "Empty",
			conf:     andromeda.GetQuotaUsageConfig{},
			lockIn:   time.Second * 1,
			maxRetry: 1,
			retryIn:  time.Millisecond * 50,
		},
		{
			name: "NotEmpty",
			conf: andromeda.GetQuotaUsageConfig{
				LockIn:   time.Second * 3,
				MaxRetry: 10,
				RetryIn:  time.Millisecond * 100,
			},
			lockIn:   time.Second * 3,
			maxRetry: 10,
			retryIn:  time.Millisecond * 100,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			conf := test.conf

			assert.Equal(t, test.lockIn, conf.GetLockIn())
			assert.Equal(t, test.maxRetry, conf.GetMaxRetry())
			assert.Equal(t, test.retryIn, conf.GetRetryIn())
		})
	}
}

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
			GetQuotaCacheParams: mockGetQuotaCacheParams,
			Next:                mockNext,
		}

		addQuotaUsage := andromeda.AddQuotaUsage(conf)

		_, ok := addQuotaUsage.(andromeda.UpdateQuotaUsage)

		assert.True(t, ok)
	})

	t.Run("ConfigWithGetQuotaUsageAndTheirConfig", func(t *testing.T) {
		mockNext := mocks.NewMockUpdateQuotaUsage(mockCtrl)
		mockGetQuotaUsage := mocks.NewMockGetQuota(mockCtrl)

		defer mockCtrl.Finish()

		conf := andromeda.AddQuotaUsageConfig{
			Cache:               mockCache,
			GetQuotaLimit:       mockGetQuotaLimit,
			GetQuotaUsage:       mockGetQuotaUsage,
			GetQuotaUsageConfig: andromeda.GetQuotaUsageConfig{},
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
			GetQuotaCacheParams: mockGetQuotaCacheParams,
			Next:                mockNext,
		}

		reduceQuotaUsage := andromeda.ReduceQuotaUsage(conf)

		_, ok := reduceQuotaUsage.(andromeda.UpdateQuotaUsage)

		assert.True(t, ok)
	})

	t.Run("ConfigWithGetQuotaUsageAndTheirConfig", func(t *testing.T) {
		mockNext := mocks.NewMockUpdateQuotaUsage(mockCtrl)
		mockGetQuotaUsage := mocks.NewMockGetQuota(mockCtrl)

		defer mockCtrl.Finish()

		conf := andromeda.ReduceQuotaUsageConfig{
			Cache:               mockCache,
			GetQuotaUsage:       mockGetQuotaUsage,
			GetQuotaUsageConfig: andromeda.GetQuotaUsageConfig{},
			GetQuotaCacheParams: mockGetQuotaCacheParams,
			Next:                mockNext,
		}

		reduceQuotaUsage := andromeda.ReduceQuotaUsage(conf)

		_, ok := reduceQuotaUsage.(andromeda.UpdateQuotaUsage)

		assert.True(t, ok)
	})
}

func BenchmarkAddQuotaUsage(b *testing.B) {
	ctx := context.TODO()
	miniRedis, err := miniredis.Run()
	assert.Nil(b, err)

	redisCache := cache.NewCacheRedis(redis.NewClient(&redis.Options{Addr: miniRedis.Addr()}))

	b.Run("WithoutGetQuotaUsage", func(b *testing.B) {
		addQuotaUsage := andromeda.AddQuotaUsage(andromeda.AddQuotaUsageConfig{
			Cache:               redisCache,
			GetQuotaCacheParams: &mockGetQuotaCacheParams{keyFormat: "quota-usage-1-%s"},
			GetQuotaLimit:       &mockGetQuota{value: 10000000},
		})

		for i := 0; i < b.N; i++ {
			_, _ = addQuotaUsage.Do(ctx, &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: 1})
		}
	})

	b.Run("WithGetQuotaUsage", func(b *testing.B) {
		addQuotaUsage := andromeda.AddQuotaUsage(andromeda.AddQuotaUsageConfig{
			Cache:               redisCache,
			GetQuotaCacheParams: &mockGetQuotaCacheParams{keyFormat: "quota-usage-2-%s"},
			GetQuotaLimit:       &mockGetQuota{value: 10000000},
			GetQuotaUsage:       &mockGetQuota{value: 0},
			GetQuotaUsageConfig: andromeda.GetQuotaUsageConfig{
				LockIn:   time.Second * 3,
				MaxRetry: 50,
				RetryIn:  time.Millisecond * 100,
			},
		})

		for i := 0; i < b.N; i++ {
			_, _ = addQuotaUsage.Do(ctx, &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: 1})
		}
	})
}

type mockGetQuota struct {
	value int64
}

func (q *mockGetQuota) Do(_ context.Context, _ *andromeda.QuotaRequest) (int64, error) {
	return q.value, nil
}

type mockGetQuotaCacheParams struct {
	keyFormat string
}

func (q *mockGetQuotaCacheParams) Do(_ context.Context, req *andromeda.QuotaRequest) (*andromeda.QuotaCacheParams, error) {
	return &andromeda.QuotaCacheParams{
		Key:        fmt.Sprintf(q.keyFormat, req.QuotaID),
		Expiration: time.Second * 30,
	}, nil
}

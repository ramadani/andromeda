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
	mockGetQuotaUsageKey := mocks.NewMockGetQuotaKey(mockCtrl)

	t.Run("PanicRequireCache", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The code did not panic")
			}

			mockCtrl.Finish()
		}()

		conf := andromeda.AddQuotaUsageConfig{}

		addQuotaUsage := andromeda.AddQuotaUsage(conf)

		_, ok := addQuotaUsage.(andromeda.UpdateQuotaUsage)

		assert.True(t, ok)
	})

	t.Run("PanicRequireGetQuotaLimit", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The code did not panic")
			}

			mockCtrl.Finish()
		}()

		conf := andromeda.AddQuotaUsageConfig{
			Cache: mockCache,
		}

		addQuotaUsage := andromeda.AddQuotaUsage(conf)

		_, ok := addQuotaUsage.(andromeda.UpdateQuotaUsage)

		assert.True(t, ok)
	})

	t.Run("PanicRequireGetQuotaUsageKey", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The code did not panic")
			}

			mockCtrl.Finish()
		}()

		conf := andromeda.AddQuotaUsageConfig{
			Cache:         mockCache,
			GetQuotaLimit: mockGetQuotaLimit,
		}

		addQuotaUsage := andromeda.AddQuotaUsage(conf)

		_, ok := addQuotaUsage.(andromeda.UpdateQuotaUsage)

		assert.True(t, ok)
	})

	t.Run("PanicRequireGetQuotaUsageExpiration", func(t *testing.T) {
		mockGetQuotaUsage := mocks.NewMockGetQuota(mockCtrl)

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The code did not panic")
			}

			mockCtrl.Finish()
		}()

		conf := andromeda.AddQuotaUsageConfig{
			Cache:            mockCache,
			GetQuotaLimit:    mockGetQuotaLimit,
			GetQuotaUsageKey: mockGetQuotaUsageKey,
			GetQuotaUsage:    mockGetQuotaUsage,
		}

		addQuotaUsage := andromeda.AddQuotaUsage(conf)

		_, ok := addQuotaUsage.(andromeda.UpdateQuotaUsage)

		assert.True(t, ok)
	})

	t.Run("ConfigWithoutNextUpdateQuotaUsage", func(t *testing.T) {
		defer mockCtrl.Finish()

		conf := andromeda.AddQuotaUsageConfig{
			Cache:            mockCache,
			GetQuotaLimit:    mockGetQuotaLimit,
			GetQuotaUsageKey: mockGetQuotaUsageKey,
		}

		addQuotaUsage := andromeda.AddQuotaUsage(conf)

		_, ok := addQuotaUsage.(andromeda.UpdateQuotaUsage)

		assert.True(t, ok)
	})

	t.Run("ConfigWithNextUpdateQuotaUsage", func(t *testing.T) {
		mockNext := mocks.NewMockUpdateQuotaUsage(mockCtrl)

		defer mockCtrl.Finish()

		conf := andromeda.AddQuotaUsageConfig{
			Cache:            mockCache,
			GetQuotaLimit:    mockGetQuotaLimit,
			GetQuotaUsageKey: mockGetQuotaUsageKey,
			Next:             mockNext,
		}

		addQuotaUsage := andromeda.AddQuotaUsage(conf)

		_, ok := addQuotaUsage.(andromeda.UpdateQuotaUsage)

		assert.True(t, ok)
	})

	t.Run("ConfigWithGetQuotaUsage", func(t *testing.T) {
		mockNext := mocks.NewMockUpdateQuotaUsage(mockCtrl)
		mockGetQuotaUsage := mocks.NewMockGetQuota(mockCtrl)
		mockGetQuotaUsageExp := mocks.NewMockGetQuotaExpiration(mockCtrl)

		defer mockCtrl.Finish()

		conf := andromeda.AddQuotaUsageConfig{
			Cache:                   mockCache,
			GetQuotaLimit:           mockGetQuotaLimit,
			GetQuotaUsage:           mockGetQuotaUsage,
			GetQuotaUsageKey:        mockGetQuotaUsageKey,
			GetQuotaUsageExpiration: mockGetQuotaUsageExp,
			Next:                    mockNext,
		}

		addQuotaUsage := andromeda.AddQuotaUsage(conf)

		_, ok := addQuotaUsage.(andromeda.UpdateQuotaUsage)

		assert.True(t, ok)
	})

	t.Run("ConfigWithGetQuotaUsageAndTheirConfig", func(t *testing.T) {
		mockNext := mocks.NewMockUpdateQuotaUsage(mockCtrl)
		mockGetQuotaUsage := mocks.NewMockGetQuota(mockCtrl)
		mockGetQuotaUsageExp := mocks.NewMockGetQuotaExpiration(mockCtrl)

		defer mockCtrl.Finish()

		conf := andromeda.AddQuotaUsageConfig{
			Cache:                   mockCache,
			GetQuotaLimit:           mockGetQuotaLimit,
			GetQuotaUsage:           mockGetQuotaUsage,
			GetQuotaUsageKey:        mockGetQuotaUsageKey,
			GetQuotaUsageExpiration: mockGetQuotaUsageExp,
			GetQuotaUsageConfig:     andromeda.GetQuotaUsageConfig{},
			Next:                    mockNext,
		}

		addQuotaUsage := andromeda.AddQuotaUsage(conf)

		_, ok := addQuotaUsage.(andromeda.UpdateQuotaUsage)

		assert.True(t, ok)
	})
}

func TestAndromedaReduceQuotaUsage(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockCache := mocks.NewMockCache(mockCtrl)
	mockGetQuotaUsageKey := mocks.NewMockGetQuotaKey(mockCtrl)

	t.Run("PanicRequireCache", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The code did not panic")
			}

			mockCtrl.Finish()
		}()

		conf := andromeda.ReduceQuotaUsageConfig{}

		reduceQuotaUsage := andromeda.ReduceQuotaUsage(conf)

		_, ok := reduceQuotaUsage.(andromeda.UpdateQuotaUsage)

		assert.True(t, ok)
	})

	t.Run("PanicRequireGetQuotaUsageKey", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The code did not panic")
			}

			mockCtrl.Finish()
		}()

		conf := andromeda.ReduceQuotaUsageConfig{
			Cache: mockCache,
		}

		reduceQuotaUsage := andromeda.ReduceQuotaUsage(conf)

		_, ok := reduceQuotaUsage.(andromeda.UpdateQuotaUsage)

		assert.True(t, ok)
	})

	t.Run("PanicRequireGetQuotaUsageExpiration", func(t *testing.T) {
		mockGetQuotaUsage := mocks.NewMockGetQuota(mockCtrl)

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The code did not panic")
			}

			mockCtrl.Finish()
		}()

		conf := andromeda.ReduceQuotaUsageConfig{
			Cache:            mockCache,
			GetQuotaUsage:    mockGetQuotaUsage,
			GetQuotaUsageKey: mockGetQuotaUsageKey,
		}

		reduceQuotaUsage := andromeda.ReduceQuotaUsage(conf)

		_, ok := reduceQuotaUsage.(andromeda.UpdateQuotaUsage)

		assert.True(t, ok)
	})

	t.Run("ConfigWithoutNextUpdateQuotaUsage", func(t *testing.T) {
		defer mockCtrl.Finish()

		conf := andromeda.ReduceQuotaUsageConfig{
			Cache:            mockCache,
			GetQuotaUsageKey: mockGetQuotaUsageKey,
		}

		reduceQuotaUsage := andromeda.ReduceQuotaUsage(conf)

		_, ok := reduceQuotaUsage.(andromeda.UpdateQuotaUsage)

		assert.True(t, ok)
	})

	t.Run("ConfigWithNextUpdateQuotaUsage", func(t *testing.T) {
		mockNext := mocks.NewMockUpdateQuotaUsage(mockCtrl)

		defer mockCtrl.Finish()

		conf := andromeda.ReduceQuotaUsageConfig{
			Cache:            mockCache,
			GetQuotaUsageKey: mockGetQuotaUsageKey,
			Next:             mockNext,
		}

		reduceQuotaUsage := andromeda.ReduceQuotaUsage(conf)

		_, ok := reduceQuotaUsage.(andromeda.UpdateQuotaUsage)

		assert.True(t, ok)
	})

	t.Run("ConfigWithGetQuotaUsage", func(t *testing.T) {
		mockNext := mocks.NewMockUpdateQuotaUsage(mockCtrl)
		mockGetQuotaUsage := mocks.NewMockGetQuota(mockCtrl)
		mockGetQuotaUsageExp := mocks.NewMockGetQuotaExpiration(mockCtrl)

		defer mockCtrl.Finish()

		conf := andromeda.ReduceQuotaUsageConfig{
			Cache:                   mockCache,
			GetQuotaUsage:           mockGetQuotaUsage,
			GetQuotaUsageKey:        mockGetQuotaUsageKey,
			GetQuotaUsageExpiration: mockGetQuotaUsageExp,
			Next:                    mockNext,
		}

		reduceQuotaUsage := andromeda.ReduceQuotaUsage(conf)

		_, ok := reduceQuotaUsage.(andromeda.UpdateQuotaUsage)

		assert.True(t, ok)
	})

	t.Run("ConfigWithGetQuotaUsageAndTheirConfig", func(t *testing.T) {
		mockNext := mocks.NewMockUpdateQuotaUsage(mockCtrl)
		mockGetQuotaUsage := mocks.NewMockGetQuota(mockCtrl)
		mockGetQuotaUsageExp := mocks.NewMockGetQuotaExpiration(mockCtrl)

		defer mockCtrl.Finish()

		conf := andromeda.ReduceQuotaUsageConfig{
			Cache:                   mockCache,
			GetQuotaUsage:           mockGetQuotaUsage,
			GetQuotaUsageKey:        mockGetQuotaUsageKey,
			GetQuotaUsageExpiration: mockGetQuotaUsageExp,
			GetQuotaUsageConfig:     andromeda.GetQuotaUsageConfig{},
			Next:                    mockNext,
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
			Cache:            redisCache,
			GetQuotaUsageKey: &mockGetQuotaKey{keyFormat: "quota-usage-1-%s"},
			GetQuotaLimit:    &mockGetQuota{value: 10000000},
		})

		for i := 0; i < b.N; i++ {
			_, _ = addQuotaUsage.Do(ctx, &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: 1})
		}
	})

	b.Run("WithGetQuotaUsage", func(b *testing.B) {
		addQuotaUsage := andromeda.AddQuotaUsage(andromeda.AddQuotaUsageConfig{
			Cache:                   redisCache,
			GetQuotaLimit:           &mockGetQuota{value: 10000000},
			GetQuotaUsage:           &mockGetQuota{value: 0},
			GetQuotaUsageKey:        &mockGetQuotaKey{keyFormat: "quota-usage-2-%s"},
			GetQuotaUsageExpiration: &mockGetQuotaExp{},
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

func BenchmarkReduceQuotaUsage(b *testing.B) {
	ctx := context.TODO()
	miniRedis, err := miniredis.Run()
	assert.Nil(b, err)

	redisCache := cache.NewCacheRedis(redis.NewClient(&redis.Options{Addr: miniRedis.Addr()}))

	reduceQuotaUsage := andromeda.ReduceQuotaUsage(andromeda.ReduceQuotaUsageConfig{
		Cache:                   redisCache,
		GetQuotaUsage:           &mockGetQuota{value: 10000000},
		GetQuotaUsageKey:        &mockGetQuotaKey{keyFormat: "quota-usage-3-%s"},
		GetQuotaUsageExpiration: &mockGetQuotaExp{},
		GetQuotaUsageConfig: andromeda.GetQuotaUsageConfig{
			LockIn:   time.Second * 3,
			MaxRetry: 50,
			RetryIn:  time.Millisecond * 100,
		},
	})

	for i := 0; i < b.N; i++ {
		_, _ = reduceQuotaUsage.Do(ctx, &andromeda.QuotaUsageRequest{QuotaID: "123", Usage: 1})
	}
}

type mockGetQuota struct {
	value int64
}

func (q *mockGetQuota) Do(_ context.Context, _ *andromeda.QuotaRequest) (int64, error) {
	return q.value, nil
}

type mockGetQuotaKey struct {
	keyFormat string
}

func (q *mockGetQuotaKey) Do(_ context.Context, req *andromeda.QuotaRequest) (string, error) {
	return fmt.Sprintf(q.keyFormat, req.QuotaID), nil
}

type mockGetQuotaExp struct{}

func (q *mockGetQuotaExp) Do(_ context.Context, req *andromeda.QuotaRequest) (time.Duration, error) {
	return time.Second * 30, nil
}

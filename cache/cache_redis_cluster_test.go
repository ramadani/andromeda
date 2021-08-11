package cache_test

import (
	"context"
	"fmt"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/ramadani/andromeda"
	"github.com/ramadani/andromeda/cache"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCacheRedisCluster(t *testing.T) {
	ctx := context.TODO()
	miniRedis, err := miniredis.Run()
	assert.Nil(t, err)

	redisCache := cache.NewCacheRedisCluster(redis.NewClusterClient(&redis.ClusterOptions{Addrs: []string{miniRedis.Addr()}}))

	t.Run("IncrByAndDecrBy", func(t *testing.T) {
		key := "123-1"

		res, err := redisCache.IncrBy(ctx, key, 1)

		assert.Equal(t, int64(1), res)
		assert.Nil(t, err)

		res, err = redisCache.DecrBy(ctx, key, 1)

		assert.Equal(t, int64(0), res)
		assert.Nil(t, err)
	})

	t.Run("SetAndGet", func(t *testing.T) {
		key := "123-2"
		ttl := 5 * time.Second

		res, err := redisCache.Set(ctx, key, "1", ttl)

		assert.Equal(t, "OK", res)
		assert.Nil(t, err)

		res, err = redisCache.Get(ctx, key)

		assert.Equal(t, "1", res)
		assert.Nil(t, err)
	})

	t.Run("ErrCacheNotFound", func(t *testing.T) {
		key := "123-3"
		res, err := redisCache.Get(ctx, key)

		assert.Equal(t, "", res)
		assert.Equal(t, andromeda.ErrCacheNotFound, err)
	})

	t.Run("SetNX", func(t *testing.T) {
		key := "123-4"
		ttl := 5 * time.Second

		res, err := redisCache.SetNX(ctx, key, "1", ttl)

		assert.True(t, res)
		assert.Nil(t, err)

		res, err = redisCache.SetNX(ctx, key, "1", ttl)

		assert.False(t, res)
		assert.Nil(t, err)
	})

	t.Run("Exists", func(t *testing.T) {
		key := "123-5"
		ttl := 5 * time.Second

		res, err := redisCache.Set(ctx, key, "1", ttl)

		assert.Equal(t, "OK", res)
		assert.Nil(t, err)

		exists, err := redisCache.Exists(ctx, key)

		assert.Equal(t, int64(1), exists)
		assert.Nil(t, err)

		exists, err = redisCache.Exists(ctx, fmt.Sprintf("test-%s", key))

		assert.Equal(t, int64(0), exists)
		assert.Nil(t, err)
	})

	t.Run("ExistsHasError", func(t *testing.T) {
		defer miniRedis.SetError("")

		key := "123-5-1"
		ttl := 5 * time.Second

		res, err := redisCache.Set(ctx, key, "1", ttl)

		assert.Equal(t, "OK", res)
		assert.Nil(t, err)

		exists, err := redisCache.Exists(ctx, fmt.Sprintf("test-%s", key))

		assert.Equal(t, int64(0), exists)
		assert.Nil(t, err)

		miniRedis.SetError("error")
		exists, err = redisCache.Exists(ctx, key)

		assert.Equal(t, int64(0), exists)
		assert.Error(t, err)
	})

	t.Run("Delete", func(t *testing.T) {
		key := "123-6"
		ttl := 5 * time.Second

		res, err := redisCache.Set(ctx, key, "1", ttl)

		assert.Equal(t, "OK", res)
		assert.Nil(t, err)

		exists, err := redisCache.Del(ctx, key)

		assert.Equal(t, int64(1), exists)
		assert.Nil(t, err)

		exists, err = redisCache.Del(ctx, key)

		assert.Equal(t, int64(0), exists)
		assert.Nil(t, err)
	})

	t.Run("DeleteHasError", func(t *testing.T) {
		defer miniRedis.SetError("")

		key := "123-6-1"
		ttl := 5 * time.Second

		res, err := redisCache.Set(ctx, key, "1", ttl)

		assert.Equal(t, "OK", res)
		assert.Nil(t, err)

		miniRedis.SetError("error")
		exists, err := redisCache.Del(ctx, key)

		assert.Equal(t, int64(0), exists)
		assert.Error(t, err)
	})
}

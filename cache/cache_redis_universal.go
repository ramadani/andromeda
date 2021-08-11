package cache

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/ramadani/andromeda"
	"time"
)

type cacheRedisUniversal struct {
	client redis.UniversalClient
}

func (c *cacheRedisUniversal) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.client.IncrBy(ctx, key, value).Result()
}

func (c *cacheRedisUniversal) DecrBy(ctx context.Context, key string, decrement int64) (int64, error) {
	return c.client.DecrBy(ctx, key, decrement).Result()
}

func (c *cacheRedisUniversal) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) (string, error) {
	return c.client.Set(ctx, key, value, expiration).Result()
}

func (c *cacheRedisUniversal) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		err = andromeda.ErrCacheNotFound
	}
	return val, err
}

func (c *cacheRedisUniversal) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return c.client.SetNX(ctx, key, value, expiration).Result()
}

func (c *cacheRedisUniversal) Exists(ctx context.Context, keys ...string) (int64, error) {
	return c.client.Exists(ctx, keys...).Result()
}

func (c *cacheRedisUniversal) Del(ctx context.Context, keys ...string) (int64, error) {
	return c.client.Del(ctx, keys...).Result()
}

// NewCacheRedisUniversal cache using redis
func NewCacheRedisUniversal(client redis.UniversalClient) andromeda.Cache {
	return &cacheRedisUniversal{client: client}
}

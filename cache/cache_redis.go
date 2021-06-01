package cache

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/ramadani/andromeda"
	"time"
)

type cacheRedis struct {
	client *redis.Client
}

func (c *cacheRedis) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.client.IncrBy(ctx, key, value).Result()
}

func (c *cacheRedis) DecrBy(ctx context.Context, key string, decrement int64) (int64, error) {
	return c.client.DecrBy(ctx, key, decrement).Result()
}

func (c *cacheRedis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) (string, error) {
	return c.client.Set(ctx, key, value, expiration).Result()
}

func (c *cacheRedis) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		err = andromeda.ErrCacheNotFound
	}
	return val, err
}

func (c *cacheRedis) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return c.client.SetNX(ctx, key, value, expiration).Result()
}

func (c *cacheRedis) Exists(ctx context.Context, keys ...string) (int64, error) {
	return c.client.Exists(ctx, keys...).Result()
}

func (c *cacheRedis) Del(ctx context.Context, keys ...string) (int64, error) {
	return c.client.Del(ctx, keys...).Result()
}

// NewCacheRedis cache using redis
func NewCacheRedis(client *redis.Client) andromeda.Cache {
	return &cacheRedis{client: client}
}

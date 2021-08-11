package cache

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/ramadani/andromeda"
	"time"
)

type cacheRedisCluster struct {
	client *redis.ClusterClient
}

func (c *cacheRedisCluster) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.client.IncrBy(ctx, key, value).Result()
}

func (c *cacheRedisCluster) DecrBy(ctx context.Context, key string, decrement int64) (int64, error) {
	return c.client.DecrBy(ctx, key, decrement).Result()
}

func (c *cacheRedisCluster) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) (string, error) {
	return c.client.Set(ctx, key, value, expiration).Result()
}

func (c *cacheRedisCluster) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		err = andromeda.ErrCacheNotFound
	}
	return val, err
}

func (c *cacheRedisCluster) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return c.client.SetNX(ctx, key, value, expiration).Result()
}

func (c *cacheRedisCluster) Exists(ctx context.Context, keys ...string) (int64, error) {
	n := int64(0)

	for _, key := range keys {
		i, err := c.client.Exists(ctx, key).Result()
		if err != nil {
			return 0, err
		}
		n += i
	}

	return n, nil
}

func (c *cacheRedisCluster) Del(ctx context.Context, keys ...string) (int64, error) {
	n := int64(0)

	for _, key := range keys {
		i, err := c.client.Del(ctx, key).Result()
		if err != nil {
			return 0, err
		}
		n += i
	}

	return n, nil
}

// NewCacheRedisCluster cache using redis cluster
func NewCacheRedisCluster(client *redis.ClusterClient) andromeda.Cache {
	return &cacheRedisCluster{client: client}
}

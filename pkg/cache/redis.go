package cache

import (
	"context"
	"time"

	"flamingo/pkg/config"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewCache)

// CacheClient Redis client
type CacheClient struct {
	rdb *redis.Client
}

// Cache abstracts Redis operations used across services.
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	Del(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, keys ...string) (int64, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	HSet(ctx context.Context, key string, values ...interface{}) error
	HGet(ctx context.Context, key, field string) (string, error)
	HGetAll(ctx context.Context, key string) (map[string]string, error)
	HDel(ctx context.Context, key string, fields ...string) error
	HIncrBy(ctx context.Context, key string, field string, incr int64) (int64, error)
	Incr(ctx context.Context, key string) (int64, error)
	Decr(ctx context.Context, key string) (int64, error)
	Close() error
}

var _ Cache = (*CacheClient)(nil)

// NewCache creates a new redis client and returns a cleanup function that
// closes the Redis connection. The func() return satisfies Wire's cleanup
// convention: when a provider returns (T, func(), error), Wire calls cleanup
// functions in reverse order on error and aggregates them.
func NewCache(cfg config.Cache) (Cache, func(), error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.Addr,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		return nil, nil, err
	}

	c := &CacheClient{rdb: rdb}
	return c, func() {
		if err := c.Close(); err != nil {
			log.Errorf("cleanup: close cache failed: %v", err)
		}
	}, nil
}

// Get gets a value
func (c *CacheClient) Get(ctx context.Context, key string) (string, error) {
	return c.rdb.Get(ctx, key).Result()
}

// Set sets a value
func (c *CacheClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.rdb.Set(ctx, key, value, expiration).Err()
}

func (c *CacheClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return c.rdb.SetNX(ctx, key, value, expiration).Result()
}

// Del deletes a key
func (c *CacheClient) Del(ctx context.Context, keys ...string) error {
	return c.rdb.Del(ctx, keys...).Err()
}

// Exists checks if key exists
func (c *CacheClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	return c.rdb.Exists(ctx, keys...).Result()
}

// Expire sets expiration time
func (c *CacheClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.rdb.Expire(ctx, key, expiration).Err()
}

// HSet sets hash field
func (c *CacheClient) HSet(ctx context.Context, key string, values ...interface{}) error {
	return c.rdb.HSet(ctx, key, values...).Err()
}

// HGet gets hash field
func (c *CacheClient) HGet(ctx context.Context, key, field string) (string, error) {
	return c.rdb.HGet(ctx, key, field).Result()
}

// HGetAll gets all hash fields
func (c *CacheClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.rdb.HGetAll(ctx, key).Result()
}

// HDel deletes hash field
func (c *CacheClient) HDel(ctx context.Context, key string, fields ...string) error {
	return c.rdb.HDel(ctx, key, fields...).Err()
}

// Incr increments
func (c *CacheClient) Incr(ctx context.Context, key string) (int64, error) {
	return c.rdb.Incr(ctx, key).Result()
}

// Decr decrements
func (c *CacheClient) Decr(ctx context.Context, key string) (int64, error) {
	return c.rdb.Decr(ctx, key).Result()
}

// HIncrBy increments the integer value of a hash field
func (c *CacheClient) HIncrBy(ctx context.Context, key string, field string, incr int64) (int64, error) {
	return c.rdb.HIncrBy(ctx, key, field, incr).Result()
}

// Close closes the connection
func (c *CacheClient) Close() error {
	return c.rdb.Close()
}

// GetClient gets the raw Redis client
func (c *CacheClient) GetClient() *redis.Client {
	return c.rdb
}

package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisClient wraps the Redis client with additional functionality
type RedisClient struct {
	client    *redis.Client
	address   string
	password  string
	db        int
	poolSize  int
	keyPrefix string
	logger    *zap.Logger
}

// Ping checks Redis connectivity
func (c *RedisClient) Ping() error {
	client := redis.NewClient(&redis.Options{
		Addr:     c.address,
		Password: c.password,
		DB:       c.db,
		PoolSize: c.poolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to ping Redis: %w", err)
	}

	c.client = client
	return nil
}

// GetKey returns a prefixed key
func (c *RedisClient) GetKey(key string) string {
	return c.keyPrefix + key
}

// Get returns a value from Redis
func (c *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, c.GetKey(key)).Result()
}

// Set sets a value in Redis with optional expiration
func (c *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, c.GetKey(key), value, expiration).Err()
}

// Delete deletes a key from Redis
func (c *RedisClient) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, c.GetKey(key)).Err()
}

// Exists checks if a key exists
func (c *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.client.Exists(ctx, c.GetKey(key)).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// TTL returns the TTL of a key
func (c *RedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.client.TTL(ctx, c.GetKey(key)).Result()
}

// Increment increments a key
func (c *RedisClient) Increment(ctx context.Context, key string) error {
	return c.client.Incr(ctx, c.GetKey(key)).Err()
}

// IncrementBy increments a key by a value
func (c *RedisClient) IncrementBy(ctx context.Context, key string, value int64) error {
	return c.client.IncrBy(ctx, c.GetKey(key), value).Err()
}

// Decrement decrements a key
func (c *RedisClient) Decrement(ctx context.Context, key string) error {
	return c.client.Decr(ctx, c.GetKey(key)).Err()
}

// SetNX sets a value if it doesn't exist
func (c *RedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return c.client.SetNX(ctx, c.GetKey(key), value, expiration).Result()
}

// HGet gets a field from a hash
func (c *RedisClient) HGet(ctx context.Context, key, field string) (string, error) {
	return c.client.HGet(ctx, c.GetKey(key), field).Result()
}

// HSet sets a field in a hash
func (c *RedisClient) HSet(ctx context.Context, key, field string, value interface{}) error {
	return c.client.HSet(ctx, c.GetKey(key), field, value).Err()
}

// HGetAll gets all fields from a hash
func (c *RedisClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.client.HGetAll(ctx, c.GetKey(key)).Result()
}

// HDel deletes a field from a hash
func (c *RedisClient) HDel(ctx context.Context, key, field string) error {
	return c.client.HDel(ctx, c.GetKey(key), field).Err()
}

// LPush pushes a value to a list
func (c *RedisClient) LPush(ctx context.Context, key string, values ...interface{}) error {
	return c.client.LPush(ctx, c.GetKey(key), values...).Err()
}

// LRange gets a range from a list
func (c *RedisClient) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return c.client.LRange(ctx, c.GetKey(key), start, stop).Result()
}

// LLen gets the length of a list
func (c *RedisClient) LLen(ctx context.Context, key string) (int64, error) {
	return c.client.LLen(ctx, c.GetKey(key)).Result()
}

// ZAdd adds a member to a sorted set
func (c *RedisClient) ZAdd(ctx context.Context, key string, score float64, member string) error {
	return c.client.ZAdd(ctx, c.GetKey(key), redis.Z{Score: score, Member: member}).Err()
}

// ZRemRangeByScore removes members by score range
func (c *RedisClient) ZRemRangeByScore(ctx context.Context, key, min, max string) error {
	return c.client.ZRemRangeByScore(ctx, c.GetKey(key), min, max).Err()
}

// ZRevRangeWithScores gets members with scores in reverse order
func (c *RedisClient) ZRevRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error) {
	return c.client.ZRevRangeWithScores(ctx, c.GetKey(key), start, stop).Result()
}

// Publish publishes a message to a channel
func (c *RedisClient) Publish(ctx context.Context, channel string, message interface{}) error {
	return c.client.Publish(ctx, c.GetKey(channel), message).Err()
}

// Subscribe subscribes to a channel
func (c *RedisClient) Subscribe(ctx context.Context, channel string) *redis.PubSub {
	return c.client.Subscribe(ctx, c.GetKey(channel))
}

// CacheRule caches a rule with TTL
func (c *RedisClient) CacheRule(ctx context.Context, ruleID string, data []byte, ttl time.Duration) error {
	return c.Set(ctx, "rule:"+ruleID, data, ttl)
}

// GetCachedRule gets a cached rule
func (c *RedisClient) GetCachedRule(ctx context.Context, ruleID string) ([]byte, error) {
	data, err := c.Get(ctx, "rule:"+ruleID)
	if err != nil {
		return nil, err
	}
	return []byte(data), nil
}

// InvalidateRuleCache removes a rule from cache
func (c *RedisClient) InvalidateRuleCache(ctx context.Context, ruleID string) error {
	return c.Delete(ctx, "rule:"+ruleID)
}

// CacheRuleset caches a ruleset with TTL
func (c *RedisClient) CacheRuleset(ctx context.Context, rulesetID string, data []byte, ttl time.Duration) error {
	return c.Set(ctx, "ruleset:"+rulesetID, data, ttl)
}

// GetCachedRuleset gets a cached ruleset
func (c *RedisClient) GetCachedRuleset(ctx context.Context, rulesetID string) ([]byte, error) {
	data, err := c.Get(ctx, "ruleset:"+rulesetID)
	if err != nil {
		return nil, err
	}
	return []byte(data), nil
}

// CacheEntity caches an entity with TTL
func (c *RedisClient) CacheEntity(ctx context.Context, entityID string, data []byte, ttl time.Duration) error {
	return c.Set(ctx, "entity:"+entityID, data, ttl)
}

// GetCachedEntity gets a cached entity
func (c *RedisClient) GetCachedEntity(ctx context.Context, entityID string) ([]byte, error) {
	data, err := c.Get(ctx, "entity:"+entityID)
	if err != nil {
		return nil, err
	}
	return []byte(data), nil
}

// IncrementViolationCounter increments the violation counter for a time window
func (c *RedisClient) IncrementViolationCounter(ctx context.Context, entityID, ruleType string, window time.Duration) (int64, error) {
	key := fmt.Sprintf("violations:%s:%s", entityID, ruleType)
	now := time.Now().Unix()

	// Use a pipeline to set expiry atomically
	pipe := c.client.Pipeline()
	pipe.Incr(ctx, c.GetKey(key))
	pipe.Expire(ctx, c.GetKey(key), window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}

	return c.client.Incr(ctx, c.GetKey(key)).Result()
}

// GetViolationCount gets the violation count for a time window
func (c *RedisClient) GetViolationCount(ctx context.Context, entityID, ruleType string) (int64, error) {
	key := fmt.Sprintf("violations:%s:%s", entityID, ruleType)
	count, err := c.client.Get(ctx, c.GetKey(key)).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return count, err
}

// TrackTransactionRate tracks the transaction rate for an entity
func (c *RedisClient) TrackTransactionRate(ctx context.Context, entityID string, window time.Duration) (int64, error) {
	key := fmt.Sprintf("tx_rate:%s", entityID)
	now := time.Now().UnixNano()

	// Use sorted set to track transactions with timestamps
	pipe := c.client.Pipeline()
	pipe.ZAdd(ctx, c.GetKey(key), redis.Z{Score: float64(now), Member: now})
	pipe.ZRemRangeByScore(ctx, c.GetKey(key), "0", fmt.Sprintf("%f", float64(now)-float64(window.Nanoseconds())))

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}

	return c.client.ZCard(ctx, c.GetKey(key)).Result()
}

// AddViolationToQueue adds a violation to a processing queue
func (c *RedisClient) AddViolationToQueue(ctx context.Context, violationID string, data []byte) error {
	return c.LPush(ctx, "violation_queue", data)
}

// GetViolationFromQueue gets a violation from the processing queue
func (c *RedisClient) GetViolationFromQueue(ctx context.Context) (string, error) {
	result, err := c.client.RPop(ctx, c.GetKey("violation_queue")).Result()
	if err == redis.Nil {
		return "", nil
	}
	return result, err
}

// CacheCheckResult caches a compliance check result
func (c *RedisClient) CacheCheckResult(ctx context.Context, txID string, data []byte, ttl time.Duration) error {
	return c.Set(ctx, "check_result:"+txID, data, ttl)
}

// GetCachedCheckResult gets a cached compliance check result
func (c *RedisClient) GetCachedCheckResult(ctx context.Context, txID string) ([]byte, error) {
	data, err := c.Get(ctx, "check_result:"+txID)
	if err != nil {
		return nil, err
	}
	return []byte(data), nil
}

// IncrementMetrics increments various metrics counters
func (c *RedisClient) IncrementMetrics(ctx context.Context, metric string) error {
	key := fmt.Sprintf("metrics:%s", metric)
	return c.Increment(ctx, key)
}

// GetMetrics gets metrics values
func (c *RedisClient) GetMetrics(ctx context.Context, metric string) (int64, error) {
	key := fmt.Sprintf("metrics:%s", metric)
	count, err := c.client.Get(ctx, c.GetKey(key)).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return count, err
}

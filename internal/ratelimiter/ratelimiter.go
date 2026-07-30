package ratelimiter

import (
	"context"
	"fmt"
	"strconv"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"edgecore/pkg/redis"
)

// RateLimitResult contains rate limiting decision details
type RateLimitResult struct {
	Allowed   bool
	Limit     int
	Remaining int
	ResetIn   time.Duration
}

// DistributedRateLimiter enforces distributed rate limits using Redis
type DistributedRateLimiter struct {
	redisClient *redis.Client
}

// NewDistributedRateLimiter instantiates a new RateLimiter
func NewDistributedRateLimiter(rc *redis.Client) *DistributedRateLimiter {
	return &DistributedRateLimiter{
		redisClient: rc,
	}
}

// Allow checks if a request is allowed for the given key and limit window
func (rl *DistributedRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (*RateLimitResult, error) {
	if limit <= 0 {
		return &RateLimitResult{Allowed: true, Limit: limit, Remaining: limit, ResetIn: 0}, nil
	}

	redisKey := fmt.Sprintf("edgecore:ratelimit:%s", key)
	now := time.Now().UnixNano()
	clearBefore := now - window.Nanoseconds()

	// If Redis client is not configured, fallback to allow
	if rl.redisClient == nil {
		return &RateLimitResult{Allowed: true, Limit: limit, Remaining: limit - 1, ResetIn: window}, nil
	}

	rdb := rl.redisClient.Raw()

	// Use Redis pipeline for sliding window ZSET rate limiting
	pipe := rdb.TxPipeline()
	pipe.ZRemRangeByScore(ctx, redisKey, "0", strconv.FormatInt(clearBefore, 10))
	countCmd := pipe.ZCard(ctx, redisKey)
	pipe.ZAdd(ctx, redisKey, goredis.Z{
		Score:  float64(now),
		Member: strconv.FormatInt(now, 10),
	})
	pipe.Expire(ctx, redisKey, window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		// Resilience: on Redis error, fail open to avoid blocking valid client traffic
		return &RateLimitResult{Allowed: true, Limit: limit, Remaining: limit - 1, ResetIn: window}, nil
	}

	currentCount := int(countCmd.Val())
	remaining := limit - currentCount - 1
	if remaining < 0 {
		remaining = 0
	}

	allowed := currentCount < limit

	return &RateLimitResult{
		Allowed:   allowed,
		Limit:     limit,
		Remaining: remaining,
		ResetIn:   window,
	}, nil
}

func (rl *DistributedRateLimiter) AllowMemoryFallback(limit int) *RateLimitResult {
	return &RateLimitResult{
		Allowed:   true,
		Limit:     limit,
		Remaining: limit,
		ResetIn:   time.Minute,
	}
}

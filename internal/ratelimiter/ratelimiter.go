package ratelimiter

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"edgecore/internal/gateway/pipeline"
	"edgecore/internal/models"
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

// RateLimiterPlugin wraps DistributedRateLimiter for gateway pipeline execution
type RateLimiterPlugin struct {
	limiter       *DistributedRateLimiter
	limit         int
	windowSeconds int
	keyStrategy   string
	headerName    string
}

func NewRateLimiterPlugin(limiter *DistributedRateLimiter) *RateLimiterPlugin {
	return &RateLimiterPlugin{
		limiter:       limiter,
		limit:         100,
		windowSeconds: 60,
		keyStrategy:   "ip",
		headerName:    "X-Client-ID",
	}
}

func (p *RateLimiterPlugin) Name() string {
	return "rate-limit"
}

func (p *RateLimiterPlugin) Init(config models.JSONMap) error {
	if config == nil {
		return nil
	}
	if l, ok := config["limit"].(float64); ok {
		p.limit = int(l)
	} else if l, ok := config["limit"].(int); ok {
		p.limit = l
	}
	if w, ok := config["window_seconds"].(float64); ok {
		p.windowSeconds = int(w)
	} else if w, ok := config["window_seconds"].(int); ok {
		p.windowSeconds = w
	}
	if strat, ok := config["key_strategy"].(string); ok && strat != "" {
		p.keyStrategy = strat
	}
	if h, ok := config["header_name"].(string); ok && h != "" {
		p.headerName = h
	}
	return nil
}

func (p *RateLimiterPlugin) ExecuteRequest(ctx *pipeline.PipelineContext) error {
	if p.limit <= 0 {
		return nil
	}

	key := p.extractKey(ctx)
	window := time.Duration(p.windowSeconds) * time.Second
	if window <= 0 {
		window = 60 * time.Second
	}

	res, err := p.limiter.Allow(ctx.Ctx, key, p.limit, window)
	if err != nil || res == nil {
		return nil
	}

	ctx.Writer.Header().Set("X-RateLimit-Limit", strconv.Itoa(res.Limit))
	ctx.Writer.Header().Set("X-RateLimit-Remaining", strconv.Itoa(res.Remaining))
	ctx.Writer.Header().Set("X-RateLimit-Reset", strconv.Itoa(int(res.ResetIn.Seconds())))

	if !res.Allowed {
		ctx.Abort(http.StatusTooManyRequests, `{"error":"Too Many Requests","message":"Rate limit quota exceeded"}`)
	}
	return nil
}

func (p *RateLimiterPlugin) ExecuteResponse(ctx *pipeline.PipelineContext) error {
	return nil
}

func (p *RateLimiterPlugin) extractKey(ctx *pipeline.PipelineContext) string {
	switch p.keyStrategy {
	case "api_key":
		if ctx.ApiKey != nil {
			return "apikey:" + ctx.ApiKey.ID.String()
		}
	case "header":
		if hVal := ctx.Request.Header.Get(p.headerName); hVal != "" {
			return "header:" + hVal
		}
	}
	return "ip:" + ctx.Request.RemoteAddr
}

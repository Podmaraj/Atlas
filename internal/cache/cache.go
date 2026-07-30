package cache

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"edgecore/internal/gateway/pipeline"
	"edgecore/internal/models"
	"edgecore/pkg/redis"
)

type CachePlugin struct {
	redisClient *redis.Client
	ttl         time.Duration
}

func NewCachePlugin(rc *redis.Client) *CachePlugin {
	return &CachePlugin{
		redisClient: rc,
		ttl:         60 * time.Second,
	}
}

func (p *CachePlugin) Name() string {
	return "response-cache"
}

func (p *CachePlugin) Init(config models.JSONMap) error {
	if ttlSec, ok := config["ttl_seconds"].(float64); ok && ttlSec > 0 {
		p.ttl = time.Duration(ttlSec) * time.Second
	}
	return nil
}

func (p *CachePlugin) ExecuteRequest(ctx *pipeline.PipelineContext) error {
	// Only cache GET requests
	if ctx.Request.Method != http.MethodGet {
		return nil
	}

	if p.redisClient == nil {
		return nil
	}

	cacheKey := p.buildCacheKey(ctx)
	cachedVal, err := p.redisClient.Get(context.Background(), cacheKey)

	if err == nil && cachedVal != "" {
		// Cache Hit
		ctx.Writer.Header().Set("X-Cache", "HIT")
		ctx.Writer.Header().Set("Content-Type", "application/json")
		ctx.Abort(http.StatusOK, cachedVal)
		return nil
	}

	ctx.Writer.Header().Set("X-Cache", "MISS")
	ctx.SetMetadata("cache_key", cacheKey)

	return nil
}

func (p *CachePlugin) ExecuteResponse(ctx *pipeline.PipelineContext) error {
	// Cache response if cache_key exists in metadata
	if cacheKeyVal, ok := ctx.GetMetadata("cache_key"); ok {
		cacheKey := cacheKeyVal.(string)
		if len(ctx.ResponseBody) > 0 && p.redisClient != nil {
			_ = p.redisClient.Set(context.Background(), cacheKey, string(ctx.ResponseBody), p.ttl)
		}
	}
	return nil
}

func (p *CachePlugin) buildCacheKey(ctx *pipeline.PipelineContext) string {
	routeID := "global"
	if ctx.Route != nil {
		routeID = ctx.Route.ID.String()
	}
	return fmt.Sprintf("edgecore:cache:%s:%s:%s", routeID, ctx.Request.Method, ctx.Request.URL.RequestURI())
}

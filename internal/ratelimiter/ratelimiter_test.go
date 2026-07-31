package ratelimiter

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"edgecore/internal/gateway/pipeline"
	"edgecore/internal/models"
)

func TestDistributedRateLimiter_MemoryFallback(t *testing.T) {
	limiter := NewDistributedRateLimiter(nil)
	res, err := limiter.Allow(context.Background(), "test-key", 10, 1*time.Minute)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !res.Allowed {
		t.Errorf("expected allowed true, got false")
	}
	if res.Limit != 10 {
		t.Errorf("expected limit 10, got %d", res.Limit)
	}
}

func TestRateLimiterPlugin_Execution(t *testing.T) {
	limiter := NewDistributedRateLimiter(nil)
	plugin := NewRateLimiterPlugin(limiter)

	err := plugin.Init(models.JSONMap{
		"limit":          float64(5),
		"window_seconds": float64(60),
		"key_strategy":   "ip",
	})
	if err != nil {
		t.Fatalf("failed to init plugin: %v", err)
	}

	if plugin.Name() != "rate-limit" {
		t.Errorf("expected plugin name 'rate-limit', got %s", plugin.Name())
	}

	req := httptest.NewRequest("GET", "/api/v1/resource", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	rec := httptest.NewRecorder()
	ctx := pipeline.NewPipelineContext(rec, req)

	err = plugin.ExecuteRequest(ctx)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	limitHeader := rec.Header().Get("X-RateLimit-Limit")
	if limitHeader != "5" {
		t.Errorf("expected X-RateLimit-Limit header '5', got '%s'", limitHeader)
	}
}

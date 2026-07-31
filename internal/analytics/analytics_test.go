package analytics

import (
	"context"
	"testing"
	"time"
)

func TestAnalyticsCollector(t *testing.T) {
	collector := NewCollector(100, 2)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	collector.Start(ctx)

	collector.Record(AccessLogRecord{
		Timestamp:  time.Now(),
		ClientIP:   "127.0.0.1",
		Method:     "GET",
		Path:       "/api/v1/health",
		StatusCode: 200,
		LatencyMs:  1.5,
		RouteID:    "route-123",
		ServiceID:  "svc-456",
	})

	time.Sleep(50 * time.Millisecond)
}

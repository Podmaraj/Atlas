package analytics

import (
	"context"
	"time"

	"go.uber.org/zap"

	"edgecore/internal/logger"
)

type AccessLogRecord struct {
	Timestamp  time.Time `json:"timestamp"`
	ClientIP   string    `json:"client_ip"`
	Method     string    `json:"method"`
	Path       string    `json:"path"`
	StatusCode int       `json:"status_code"`
	LatencyMs  float64   `json:"latency_ms"`
	RouteID    string    `json:"route_id"`
	ServiceID  string    `json:"service_id"`
	TenantID   string    `json:"tenant_id"`
	UserAgent  string    `json:"user_agent"`
}

type Collector struct {
	logChan chan AccessLogRecord
	workers int
}

func NewCollector(bufferSize int, workers int) *Collector {
	if bufferSize <= 0 {
		bufferSize = 10000
	}
	if workers <= 0 {
		workers = 4
	}
	return &Collector{
		logChan: make(chan AccessLogRecord, bufferSize),
		workers: workers,
	}
}

// Start launches worker pool reading access log records from queue asynchronously
func (c *Collector) Start(ctx context.Context) {
	for i := 0; i < c.workers; i++ {
		go func(workerID int) {
			for {
				select {
				case <-ctx.Done():
					return
				case record, ok := <-c.logChan:
					if !ok {
						return
					}
					c.processRecord(record)
				}
			}
		}(i)
	}

	logger.Info("Asynchronous Access Log Analytics Collector started", zap.Int("workers", c.workers))
}

// Record dispatches an access log entry to non-blocking worker queue
func (c *Collector) Record(record AccessLogRecord) {
	select {
	case c.logChan <- record:
	default:
		// Queue full - drop log entry to protect Data Plane latency
		logger.Warn("Analytics log buffer queue full - record dropped")
	}
}

func (c *Collector) processRecord(rec AccessLogRecord) {
	logger.Info("HTTP Access Log",
		zap.String("client_ip", rec.ClientIP),
		zap.String("method", rec.Method),
		zap.String("path", rec.Path),
		zap.Int("status", rec.StatusCode),
		zap.Float64("latency_ms", rec.LatencyMs),
		zap.String("route_id", rec.RouteID),
		zap.String("service_id", rec.ServiceID),
	)
}

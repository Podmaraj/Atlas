package logger_test

import (
	"context"
	"testing"

	"edgecore/internal/config"
	"edgecore/internal/logger"
)

func TestInitLoggerAndContext(t *testing.T) {
	cfg := config.LoggerConfig{
		Level:       "debug",
		Format:      "console",
		Development: true,
	}

	err := logger.InitLogger(cfg)
	if err != nil {
		t.Fatalf("failed to init logger: %v", err)
	}

	ctx := context.WithValue(context.Background(), logger.RequestIDKey, "req-12345")
	ctx = context.WithValue(ctx, logger.TenantIDKey, "tenant-abc")

	l := logger.WithContext(ctx)
	if l == nil {
		t.Fatal("expected logger with context to not be nil")
	}

	logger.Info("Logger test successfully executed")
}

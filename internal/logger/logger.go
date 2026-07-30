package logger

import (
	"context"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"edgecore/internal/config"
)

type contextKey string

const (
	RequestIDKey contextKey = "request_id"
	TraceIDKey   contextKey = "trace_id"
	TenantIDKey  contextKey = "tenant_id"
	UserIDKey    contextKey = "user_id"
)

var (
	globalLogger *zap.Logger
	sugarLogger  *zap.SugaredLogger
	once         sync.Once
)

// InitLogger initializes the global Uber Zap logger using application config
func InitLogger(cfg config.LoggerConfig) error {
	var err error
	once.Do(func() {
		var zapCfg zap.Config

		if cfg.Development {
			zapCfg = zap.NewDevelopmentConfig()
			zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		} else {
			zapCfg = zap.NewProductionConfig()
			zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		}

		var level zapcore.Level
		if err = level.UnmarshalText([]byte(cfg.Level)); err != nil {
			level = zapcore.InfoLevel
		}
		zapCfg.Level = zap.NewAtomicLevelAt(level)

		if cfg.Format == "console" {
			zapCfg.Encoding = "console"
		} else {
			zapCfg.Encoding = "json"
		}

		globalLogger, err = zapCfg.Build(zap.AddCallerSkip(1))
		if err != nil {
			return
		}
		sugarLogger = globalLogger.Sugar()
	})

	return err
}

// L returns the global raw Zap logger
func L() *zap.Logger {
	if globalLogger == nil {
		// Fallback to basic logger if uninitialized
		globalLogger, _ = zap.NewProduction()
		sugarLogger = globalLogger.Sugar()
	}
	return globalLogger
}

// S returns the global sugared logger
func S() *zap.SugaredLogger {
	if sugarLogger == nil {
		_ = L()
	}
	return sugarLogger
}

// WithContext returns a logger enriched with fields extracted from context.Context
func WithContext(ctx context.Context) *zap.Logger {
	l := L()
	if ctx == nil {
		return l
	}

	var fields []zap.Field

	if reqID, ok := ctx.Value(RequestIDKey).(string); ok && reqID != "" {
		fields = append(fields, zap.String("request_id", reqID))
	}
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok && traceID != "" {
		fields = append(fields, zap.String("trace_id", traceID))
	}
	if tenantID, ok := ctx.Value(TenantIDKey).(string); ok && tenantID != "" {
		fields = append(fields, zap.String("tenant_id", tenantID))
	}
	if userID, ok := ctx.Value(UserIDKey).(string); ok && userID != "" {
		fields = append(fields, zap.String("user_id", userID))
	}

	if len(fields) > 0 {
		return l.With(fields...)
	}

	return l
}

// Helper logging functions
func Info(msg string, fields ...zap.Field) {
	L().Info(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	L().Error(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	L().Debug(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	L().Warn(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	L().Fatal(msg, fields...)
}

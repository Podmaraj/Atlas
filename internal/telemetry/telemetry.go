package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const TracerName = "edgecore-gateway"

func InitTelemetry() (func(context.Context) error, error) {
	// Set global W3C TraceContext propagator for distributed headers (traceparent)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	shutdown := func(ctx context.Context) error {
		return nil
	}

	return shutdown, nil
}

// Tracer returns named tracer instance for creating spans
func Tracer() trace.Tracer {
	return otel.Tracer(TracerName)
}

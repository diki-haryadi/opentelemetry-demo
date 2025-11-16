package otellib

import (
	"context"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// HTTPMiddleware wraps an http.Handler with OpenTelemetry instrumentation
func (t *Telemetry) HTTPMiddleware(handler http.Handler, operation string) http.Handler {
	return otelhttp.NewHandler(handler, operation)
}

// HTTPClient returns an instrumented HTTP client
func (t *Telemetry) HTTPClient() *http.Client {
	return &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
}

// RecordMetrics records HTTP request metrics
func (t *Telemetry) RecordMetrics(ctx context.Context, method, endpoint string, statusCode int) {
	if t.requestCounter == nil {
		return
	}

	attrs := metric.WithAttributes(
		attribute.String("http.method", method),
		attribute.String("http.route", endpoint),
		attribute.Int("http.status_code", statusCode),
	)

	t.requestCounter.Add(ctx, 1, attrs)
}

// StartSpan starts a new span with the given name
func (t *Telemetry) StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	if t.tracer == nil {
		return ctx, trace.SpanFromContext(ctx)
	}
	return t.tracer.Start(ctx, name)
}

// SpanFromContext extracts a span from the context
func (t *Telemetry) SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// SetSpanAttributes sets attributes on the span from context
func (t *Telemetry) SetSpanAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attrs...)
}

// RecordSpanError records an error on the span from context
func (t *Telemetry) RecordSpanError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
}

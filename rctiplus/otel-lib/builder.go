package otellib

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// TelemetryBuilder is the builder for creating OpenTelemetry components
type TelemetryBuilder struct {
	config *Config
}

// NewBuilder creates a new TelemetryBuilder with default configuration
func NewBuilder() *TelemetryBuilder {
	return &TelemetryBuilder{
		config: DefaultConfig(),
	}
}

// WithConfig sets a custom configuration
func (b *TelemetryBuilder) WithConfig(config *Config) *TelemetryBuilder {
	b.config = config
	return b
}

// Build creates and initializes all OpenTelemetry components
func (b *TelemetryBuilder) Build(ctx context.Context) (*Telemetry, error) {
	// Build resource
	res, err := b.config.buildResource(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	telemetry := &Telemetry{
		config:   b.config,
		resource: res,
	}

	// Initialize tracer if enabled
	if b.config.TraceEnabled {
		if err := telemetry.initTracer(ctx, res); err != nil {
			return nil, fmt.Errorf("failed to initialize tracer: %w", err)
		}
	}

	// Initialize meter if enabled
	if b.config.MetricEnabled {
		if err := telemetry.initMeter(ctx, res); err != nil {
			return nil, fmt.Errorf("failed to initialize meter: %w", err)
		}
	}

	// Setup propagators
	telemetry.setupPropagators()

	return telemetry, nil
}

// Telemetry holds all OpenTelemetry components
type Telemetry struct {
	config   *Config
	resource *resource.Resource

	tracerProvider *sdktrace.TracerProvider
	tracer         trace.Tracer

	meterProvider *sdkmetric.MeterProvider
	meter         metric.Meter

	// Pre-built metrics
	requestCounter  metric.Int64Counter
	requestDuration metric.Float64Histogram
	activeRequests  metric.Int64UpDownCounter
}

// initTracer initializes the tracer provider and tracer
func (t *Telemetry) initTracer(ctx context.Context, res *resource.Resource) error {
	endpoint := t.config.TraceEndpoint
	if endpoint == "" {
		endpoint = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	}

	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithInsecure(),
	}
	if endpoint != "" {
		opts = append(opts, otlptracegrpc.WithEndpoint(endpoint))
	}

	traceExporter, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		return fmt.Errorf("failed to create trace exporter: %w", err)
	}

	t.tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(t.tracerProvider)
	t.tracer = t.tracerProvider.Tracer(t.config.ServiceName)

	return nil
}

// initMeter initializes the meter provider and meter
func (t *Telemetry) initMeter(ctx context.Context, res *resource.Resource) error {
	endpoint := t.config.MetricEndpoint
	if endpoint == "" {
		endpoint = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	}

	opts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithInsecure(),
	}
	if endpoint != "" {
		opts = append(opts, otlpmetricgrpc.WithEndpoint(endpoint))
	}

	metricExporter, err := otlpmetricgrpc.New(ctx, opts...)
	if err != nil {
		return fmt.Errorf("failed to create metric exporter: %w", err)
	}

	reader := sdkmetric.NewPeriodicReader(metricExporter,
		sdkmetric.WithInterval(t.config.MetricInterval))

	optsProvider := []sdkmetric.Option{
		sdkmetric.WithReader(reader),
		sdkmetric.WithResource(res),
	}

	// Add custom view if histogram boundaries are configured
	if len(t.config.HistogramBoundaries) > 0 {
		optsProvider = append(optsProvider,
			sdkmetric.WithView(t.config.buildMetricView()))
	}

	t.meterProvider = sdkmetric.NewMeterProvider(optsProvider...)
	otel.SetMeterProvider(t.meterProvider)

	t.meter = t.meterProvider.Meter(t.config.ServiceName)

	// Create default metrics
	if err := t.createDefaultMetrics(); err != nil {
		return fmt.Errorf("failed to create default metrics: %w", err)
	}

	return nil
}

// setupPropagators sets up the text map propagators
func (t *Telemetry) setupPropagators() {
	var propagators []propagation.TextMapPropagator

	for _, propName := range t.config.Propagators {
		switch propName {
		case "tracecontext":
			propagators = append(propagators, propagation.TraceContext{})
		case "baggage":
			propagators = append(propagators, propagation.Baggage{})
		}
	}

	if len(propagators) > 0 {
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagators...))
	}
}

// createDefaultMetrics creates the default HTTP metrics
func (t *Telemetry) createDefaultMetrics() error {
	var err error

	t.requestCounter, err = t.meter.Int64Counter(
		"http.server.request.count",
		metric.WithDescription("Number of HTTP requests"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return err
	}

	t.requestDuration, err = t.meter.Float64Histogram(
		"http.server.request.duration",
		metric.WithDescription("Duration of HTTP requests"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return err
	}

	t.activeRequests, err = t.meter.Int64UpDownCounter(
		"http.server.active_requests",
		metric.WithDescription("Number of active HTTP requests"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return err
	}

	return nil
}

// Shutdown gracefully shuts down all OpenTelemetry components
func (t *Telemetry) Shutdown(ctx context.Context) error {
	var errs []error

	if t.tracerProvider != nil {
		if err := t.tracerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if t.meterProvider != nil {
		if err := t.meterProvider.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during shutdown: %v", errs)
	}

	return nil
}

// Tracer returns the tracer instance
func (t *Telemetry) Tracer() trace.Tracer {
	return t.tracer
}

// Meter returns the meter instance
func (t *Telemetry) Meter() metric.Meter {
	return t.meter
}

// RequestCounter returns the request counter metric
func (t *Telemetry) RequestCounter() metric.Int64Counter {
	return t.requestCounter
}

// RequestDuration returns the request duration histogram metric
func (t *Telemetry) RequestDuration() metric.Float64Histogram {
	return t.requestDuration
}

// ActiveRequests returns the active requests up-down counter metric
func (t *Telemetry) ActiveRequests() metric.Int64UpDownCounter {
	return t.activeRequests
}


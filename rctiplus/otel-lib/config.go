package otellib

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// Config holds the configuration for OpenTelemetry initialization
type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	Attributes     []attribute.KeyValue

	// Trace configuration
	TraceEnabled bool
	TraceEndpoint string

	// Metric configuration
	MetricEnabled      bool
	MetricEndpoint     string
	MetricInterval     time.Duration
	HistogramBoundaries []float64

	// Propagation
	Propagators []string
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		ServiceName:    "unknown-service",
		ServiceVersion: "1.0.0",
		Environment:    "production",
		Attributes:     []attribute.KeyValue{},

		TraceEnabled:  true,
		TraceEndpoint: "", // Will use OTEL_EXPORTER_OTLP_ENDPOINT env var

		MetricEnabled:      true,
		MetricEndpoint:     "", // Will use OTEL_EXPORTER_OTLP_ENDPOINT env var
		MetricInterval:     15 * time.Second,
		HistogramBoundaries: []float64{0, 5, 10, 25, 50, 75, 100, 250, 500, 750, 1000, 2500, 5000, 7500, 10000},

		Propagators: []string{"tracecontext", "baggage"},
	}
}

// WithServiceName sets the service name
func (c *Config) WithServiceName(name string) *Config {
	c.ServiceName = name
	return c
}

// WithServiceVersion sets the service version
func (c *Config) WithServiceVersion(version string) *Config {
	c.ServiceVersion = version
	return c
}

// WithEnvironment sets the environment
func (c *Config) WithEnvironment(env string) *Config {
	c.Environment = env
	return c
}

// WithAttributes adds custom attributes
func (c *Config) WithAttributes(attrs ...attribute.KeyValue) *Config {
	c.Attributes = append(c.Attributes, attrs...)
	return c
}

// WithTraceEnabled enables or disables tracing
func (c *Config) WithTraceEnabled(enabled bool) *Config {
	c.TraceEnabled = enabled
	return c
}

// WithTraceEndpoint sets the trace endpoint
func (c *Config) WithTraceEndpoint(endpoint string) *Config {
	c.TraceEndpoint = endpoint
	return c
}

// WithMetricEnabled enables or disables metrics
func (c *Config) WithMetricEnabled(enabled bool) *Config {
	c.MetricEnabled = enabled
	return c
}

// WithMetricEndpoint sets the metric endpoint
func (c *Config) WithMetricEndpoint(endpoint string) *Config {
	c.MetricEndpoint = endpoint
	return c
}

// WithMetricInterval sets the metric export interval
func (c *Config) WithMetricInterval(interval time.Duration) *Config {
	c.MetricInterval = interval
	return c
}

// WithHistogramBoundaries sets custom histogram boundaries
func (c *Config) WithHistogramBoundaries(boundaries []float64) *Config {
	c.HistogramBoundaries = boundaries
	return c
}

// WithPropagators sets the propagation formats
func (c *Config) WithPropagators(propagators ...string) *Config {
	c.Propagators = propagators
	return c
}

// buildResource creates a resource from the configuration
func (c *Config) buildResource(ctx context.Context) (*resource.Resource, error) {
	attrs := []attribute.KeyValue{
		semconv.ServiceNameKey.String(c.ServiceName),
		semconv.ServiceVersionKey.String(c.ServiceVersion),
		attribute.String("environment", c.Environment),
	}
	attrs = append(attrs, c.Attributes...)

	return resource.New(
		ctx,
		resource.WithAttributes(attrs...),
	)
}

// buildMetricView creates a metric view for HTTP request duration
func (c *Config) buildMetricView() sdkmetric.View {
	return sdkmetric.NewView(
		sdkmetric.Instrument{Name: "http.server.request.duration"},
		sdkmetric.Stream{
			Aggregation: sdkmetric.AggregationExplicitBucketHistogram{
				Boundaries: c.HistogramBoundaries,
			},
		},
	)
}


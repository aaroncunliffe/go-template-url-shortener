package telemetry

import (
	"context"
	"errors"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	otelprometheus "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Telemetry struct {
	ServiceName  string
	ServiceBuild string

	// Providers
	metricsProvider *sdkmetric.MeterProvider
	traceProvider   *sdktrace.TracerProvider

	// Metrics
	httpRequestsTotal   metric.Int64Counter
	httpRequestDuration metric.Float64Histogram
	eventsTotal         metric.Int64Counter
}

type Config struct {
	ServiceName  string
	ServiceBuild string

	TraceEndpoint    string
	TraceProbability float64
}

func New(ctx context.Context, config Config) (*Telemetry, error) {
	// Tracing
	traceExporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(config.TraceEndpoint))
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String("service.name", config.ServiceName),
			attribute.String("service.build", config.ServiceBuild),
		),
	)
	if err != nil {
		return nil, err
	}

	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(config.TraceProbability)),
	)
	otel.SetTracerProvider(traceProvider)

	// Parameters useful especially in a distributed workload - WC3 standard
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Metrics
	metricsExporter, err := otelprometheus.New()
	if err != nil {
		return nil, err
	}

	metricsProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(metricsExporter),
		sdkmetric.WithView(latencyHistogramView("http_request_duration_seconds")),
		sdkmetric.WithView(latencyHistogramView("db.client.operation.duration")),
	)
	otel.SetMeterProvider(metricsProvider)
	meter := metricsProvider.Meter(config.ServiceName)

	httpRequestsTotal, err := meter.Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total number of HTTP requests handled."),
	)
	if err != nil {
		return nil, err
	}

	httpRequestDuration, err := meter.Float64Histogram(
		"http_request_duration_seconds",
		metric.WithDescription("Duration of HTTP requests in seconds."),
	)
	if err != nil {
		return nil, err
	}

	eventsTotal, err := meter.Int64Counter(
		"events_total",
		metric.WithDescription("Total count of events by type and result."),
	)
	if err != nil {
		return nil, err
	}

	return &Telemetry{
		ServiceName:  config.ServiceName,
		ServiceBuild: config.ServiceBuild,

		metricsProvider: metricsProvider,
		traceProvider:   traceProvider,

		httpRequestsTotal:   httpRequestsTotal,
		httpRequestDuration: httpRequestDuration,
		eventsTotal:         eventsTotal,
	}, nil
}

func (t *Telemetry) Shutdown(ctx context.Context) error {
	var errs []error

	if t.metricsProvider != nil {
		errs = append(errs, t.metricsProvider.Shutdown(ctx))
	}
	if t.traceProvider != nil {
		errs = append(errs, t.traceProvider.Shutdown(ctx))
	}

	return errors.Join(errs...)
}

func (t *Telemetry) MetricsHandler() http.Handler {
	return promhttp.Handler()
}

func (t *Telemetry) RecordHTTPRequest(ctx context.Context, method, route, status string, duration float64) {
	attrs := []attribute.KeyValue{
		attribute.String("method", method),
		attribute.String("path", route),
		attribute.String("status", status),
	}
	t.httpRequestsTotal.Add(ctx, 1, metric.WithAttributes(attrs...))
	t.httpRequestDuration.Record(ctx, duration, metric.WithAttributes(attrs...))
}

func (t *Telemetry) RecordEvent(ctx context.Context, event string, err error) {
	t.eventsTotal.Add(ctx, 1, metric.WithAttributes(
		attribute.String("event", event),
		attribute.String("result", resultLabel(err)),
	))
}

// Simple function to extract prometheus label safe 'code' from error
func resultLabel(err error) string {
	if err == nil {
		return "ok"
	}
	var c interface{ Code() string }
	if errors.As(err, &c) {
		return c.Code()
	}
	return "error"
}

// Simple wrapper to help configure metrics with consistent granular histogram buckets
func latencyHistogramView(name string) sdkmetric.View {
	return sdkmetric.NewView(
		sdkmetric.Instrument{Name: name},
		sdkmetric.Stream{
			Aggregation: sdkmetric.AggregationExplicitBucketHistogram{
				Boundaries: prometheus.DefBuckets,
			},
		},
	)
}

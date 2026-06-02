package telemetry

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelprometheus "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

type Telemetry struct {
	// Providers
	metricsProvider *sdkmetric.MeterProvider

	// Metrics
	httpRequestsTotal   metric.Int64Counter
	httpRequestDuration metric.Float64Histogram
	eventsTotal         metric.Int64Counter
}

func Setup(ctx context.Context, logger *slog.Logger, name string) (*Telemetry, error) {
	exporter, err := otelprometheus.New()
	if err != nil {
		return nil, err
	}

	// Metrics
	metricsProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(exporter),
		sdkmetric.WithView(latencyHistogramView("http_request_duration_seconds")),
		sdkmetric.WithView(latencyHistogramView("db.client.operation.duration")),
	)
	otel.SetMeterProvider(metricsProvider)
	meter := metricsProvider.Meter(name)

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

	logger.Info("telemetry setup complete")
	return &Telemetry{
		metricsProvider: metricsProvider,

		httpRequestsTotal:   httpRequestsTotal,
		httpRequestDuration: httpRequestDuration,
		eventsTotal:         eventsTotal,
	}, nil
}

func (t *Telemetry) Shutdown(ctx context.Context) error {
	if t.metricsProvider != nil {
		return t.metricsProvider.Shutdown(ctx)
	}
	return nil
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

// Simple wrapper to help configur metrics with granular histogram buckets
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

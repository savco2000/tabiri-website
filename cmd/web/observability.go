package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.opentelemetry.io/otel/trace"
)

const instrumentationName = "website.tabirianalytics.com/cmd/web"

type telemetry struct {
	tracer trace.Tracer

	requests         metric.Int64Counter
	requestDuration  metric.Float64Histogram
	inFlight         metric.Int64UpDownCounter
	templateFailures metric.Int64Counter
	databaseDuration metric.Float64Histogram
	databaseFailures metric.Int64Counter
	leadOutcomes     metric.Int64Counter
	turnstileCalls   metric.Int64Counter
	turnstileLatency metric.Float64Histogram

	tracerProvider *sdktrace.TracerProvider
	meterProvider  *sdkmetric.MeterProvider
}

func newTelemetry(ctx context.Context, cfg *config) (*telemetry, error) {
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("tabiri-website"),
			attribute.String("deployment.environment.name", cfg.environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create telemetry resource: %w", err)
	}

	traceOptions := []sdktrace.TracerProviderOption{
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(
			sdktrace.TraceIDRatioBased(cfg.traceSampleRatio),
		)),
	}
	metricOptions := []sdkmetric.Option{sdkmetric.WithResource(res)}
	if cfg.telemetryEnabled {
		traceExporter, err := otlptracehttp.New(ctx)
		if err != nil {
			return nil, fmt.Errorf("create OTLP trace exporter: %w", err)
		}
		metricExporter, err := otlpmetrichttp.New(ctx)
		if err != nil {
			return nil, fmt.Errorf("create OTLP metric exporter: %w", err)
		}
		traceOptions = append(traceOptions, sdktrace.WithBatcher(traceExporter))
		metricOptions = append(metricOptions, sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(metricExporter),
		))
	}

	tracerProvider := sdktrace.NewTracerProvider(traceOptions...)
	meterProvider := sdkmetric.NewMeterProvider(metricOptions...)
	otel.SetTracerProvider(tracerProvider)
	otel.SetMeterProvider(meterProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	meter := meterProvider.Meter(instrumentationName)
	t := &telemetry{
		tracer:         tracerProvider.Tracer(instrumentationName),
		tracerProvider: tracerProvider,
		meterProvider:  meterProvider,
	}
	if t.requests, err = meter.Int64Counter("http.server.request.count"); err != nil {
		return nil, err
	}
	if t.requestDuration, err = meter.Float64Histogram(
		"http.server.request.duration",
		metric.WithUnit("s"),
	); err != nil {
		return nil, err
	}
	if t.inFlight, err = meter.Int64UpDownCounter("http.server.active_requests"); err != nil {
		return nil, err
	}
	if t.templateFailures, err = meter.Int64Counter("template.render.failure.count"); err != nil {
		return nil, err
	}
	if t.databaseDuration, err = meter.Float64Histogram(
		"database.operation.duration",
		metric.WithUnit("s"),
	); err != nil {
		return nil, err
	}
	if t.databaseFailures, err = meter.Int64Counter("database.operation.failure.count"); err != nil {
		return nil, err
	}
	if t.leadOutcomes, err = meter.Int64Counter("lead.submission.count"); err != nil {
		return nil, err
	}
	if t.turnstileCalls, err = meter.Int64Counter("turnstile.request.count"); err != nil {
		return nil, err
	}
	if t.turnstileLatency, err = meter.Float64Histogram(
		"turnstile.request.duration",
		metric.WithUnit("s"),
	); err != nil {
		return nil, err
	}
	return t, nil
}

func (t *telemetry) shutdown(ctx context.Context) error {
	var errs []error
	if err := t.meterProvider.Shutdown(ctx); err != nil {
		errs = append(errs, err)
	}
	if err := t.tracerProvider.Shutdown(ctx); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func (t *telemetry) recordHTTPRequest(
	ctx context.Context,
	method string,
	route string,
	status int,
	duration time.Duration,
) {
	attrs := metric.WithAttributes(
		attribute.String("http.request.method", method),
		attribute.String("http.route", route),
		attribute.String("http.response.status_code", strconv.Itoa(status)),
	)
	t.requests.Add(ctx, 1, attrs)
	t.requestDuration.Record(ctx, duration.Seconds(), attrs)
}

func (t *telemetry) recordDatabase(
	ctx context.Context,
	operation string,
	duration time.Duration,
	err error,
) {
	attrs := metric.WithAttributes(attribute.String("db.operation.name", operation))
	t.databaseDuration.Record(ctx, duration.Seconds(), attrs)
	if err != nil {
		t.databaseFailures.Add(ctx, 1, attrs)
	}
}

func (t *telemetry) recordLeadOutcome(ctx context.Context, form, outcome string) {
	t.leadOutcomes.Add(ctx, 1, metric.WithAttributes(
		attribute.String("lead.form", form),
		attribute.String("lead.outcome", outcome),
	))
}

func (t *telemetry) recordTurnstile(ctx context.Context, outcome string, duration time.Duration) {
	attrs := metric.WithAttributes(attribute.String("turnstile.outcome", outcome))
	t.turnstileCalls.Add(ctx, 1, attrs)
	t.turnstileLatency.Record(ctx, duration.Seconds(), attrs)
}

func telemetryConfigured() bool {
	return strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")) != ""
}

package tracing

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	defaultSampleRatio float64 = 0.01
)

// Init returns a newly configured tracer
func Init(serviceName, host string) (trace.Tracer, error) {
	ratio := defaultSampleRatio
	if val, ok := os.LookupEnv("OTEL_SAMPLE_RATIO"); ok {
		ratio, _ = strconv.ParseFloat(val, 64)
		if ratio > 1 {
			ratio = 1.0
		}
	}

	// Get OTEL endpoint from environment variable or use host parameter
	endpoint := host
	if val, ok := os.LookupEnv("OTEL_EXPORTER_OTLP_ENDPOINT"); ok {
		endpoint = val
	}

	fmt.Printf("OpenTelemetry client: adjusted sample ratio %f, endpoint: %s\n", ratio, endpoint)

	// Create OTLP HTTP exporter
	ctx := context.Background()
	client := otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
	)
	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create OTLP exporter: %v\n", err)
		return nil, err
	}

	// Create resource with service name
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create resource: %v\n", err)
		return nil, err
	}

	// Create tracer provider with sampling
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(ratio)),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Set global propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	fmt.Println("OpenTelemetry tracer initialized successfully")
	return tp.Tracer(serviceName), nil
}

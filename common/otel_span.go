package common

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"log"
	"strings"
)

var tracerPrefix = "app.heero/"
var tracerProvider *sdktrace.TracerProvider

type defaultHeeroSampler struct{}

func (hs *defaultHeeroSampler) ShouldSample(params sdktrace.SamplingParameters) sdktrace.SamplingResult {
	if strings.HasPrefix(params.Name, tracerPrefix) {
		return sdktrace.SamplingResult{
			Decision:   sdktrace.RecordAndSample,
			Attributes: params.Attributes,
		}
	}
	return sdktrace.SamplingResult{Decision: sdktrace.Drop}
}

func (hs *defaultHeeroSampler) Description() string {
	return "heeroSampler"
}

func NewTraceProvider(exp sdktrace.SpanExporter, serviceName string) *sdktrace.TracerProvider {
	r, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		),
	)

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(r),
	)
}

type HeeroTracer struct {
	trace.Tracer
	prefix string
}

func NewHeeroTracer(ctx context.Context, serviceName string) (HeeroTracer, error) {
	exp, err := otlptrace.New(ctx, otlptracehttp.NewClient())
	if err != nil {

		log.Fatalf("failed to initialize exporter: %v", err)
		return HeeroTracer{}, err
	}
	if tracerProvider == nil {
		tracerProvider = NewTraceProvider(exp, serviceName)
	}

	otel.SetTracerProvider(tracerProvider)

	tracer := tracerProvider.Tracer(serviceName)
	return HeeroTracer{
		Tracer: tracer,
		prefix: tracerPrefix + serviceName + "/",
	}, nil
}

func ShutdownTracerProvider(ctx context.Context) error {
	return tracerProvider.Shutdown(ctx)
}

func (h HeeroTracer) Start(ctx context.Context, name string, _ ...trace.SpanStartOption) (context.Context, trace.Span) {
	return h.Tracer.Start(ctx, h.prefix+name)
}

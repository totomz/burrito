package common

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// StartOpenTelemetryPrometheus configure and start OpenTelemetry with a Prometheus exporter
func StartOpenTelemetryPrometheus(serviceName string) func() {
	slog.Info("initializing telemetry")

	res, _ := resource.Merge(resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceNamespace("heero"),
			semconv.ServiceName(serviceName),
			semconv.DeploymentEnvironment(string(GetEnvironment())),
		))

	metricExporter, err := prometheus.New()
	if err != nil {
		log.Fatal(err)
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(metricExporter),
	)

	go serveMetrics()

	otel.SetMeterProvider(meterProvider)

	return func() {
		if err := meterProvider.Shutdown(context.Background()); err != nil {
			slog.Error("error stopping the meter provider", "error", err)
		}
	}
}

func serveMetrics() {
	slog.Info("serving metrics at localhost:2223/metrics")
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(":2223", nil) //nolint:gosec // Ignoring G114: Use of net/http serve function that has no support for setting timeouts.
	if err != nil {
		panic(fmt.Sprintf("error serving http: %v", err))
	}
}

func MustGetCwd() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return wd
}

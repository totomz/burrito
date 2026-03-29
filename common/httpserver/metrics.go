package httpserver

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"time"
)

var (
	DURATION_SECONDS_BUCKET = []float64{0.0, 0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1.0, 2.5, 3.0, 5.0, 7.5, 10.0}

	// mHttpServerDurationS http request duration in seconds
	mHttpServerDurationS metric.Float64Histogram
)

func initMetrics() {
	meter := otel.Meter("common-httpserver")
	var err error

	mHttpServerDurationS, err = meter.Float64Histogram(
		"http.server.duration",
		metric.WithDescription("Number of HTTP request."),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(DURATION_SECONDS_BUCKET...),
	)
	if err != nil {
		stderr.Printf("can't initialize metric [http.server.duration]: %v", err)
	}
}

func sinceInMilliseconds(startTime time.Time) float64 {
	return float64(time.Since(startTime).Nanoseconds()) / 1e6
}

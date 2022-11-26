package httpserver

import (
	"contrib.go.opencensus.io/exporter/prometheus"
	"contrib.go.opencensus.io/exporter/zipkin"
	openzipkin "github.com/openzipkin/zipkin-go"
	zipkinHTTP "github.com/openzipkin/zipkin-go/reporter/http"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.opencensus.io/trace"
	"log"
	"net/http"
	"time"
)

type StandardMetrics struct {
	Namespace string
	// ConstLabels will be set as labels on all views.
	ConstLabels map[string]string
}

var (
	MetricApiLatency = stats.Float64("httpserver/api/latency", "The latency in milliseconds per REPL loop", stats.UnitMilliseconds)
	MetricApiStatus  = stats.Int64("httpserver/api/status", "The distribution of line lengths", stats.UnitDimensionless)

	KeyMethod, _ = tag.NewKey("method")
	KeyStatus, _ = tag.NewKey("status")
	KeyPath, _   = tag.NewKey("path")

	LatencyView = &view.View{
		Measure:     MetricApiLatency,
		Description: "The distribution of the latencies",
		// Latency in buckets:
		// [>=0ms, >=25ms, >=50ms, >=75ms, >=100ms, >=200ms, >=400ms, >=600ms, >=800ms, >=1s, >=2s, >=4s, >=6s]
		Aggregation: view.Distribution(0, 25, 50, 75, 100, 200, 400, 600, 800, 1000, 2000, 4000, 6000),
		TagKeys:     []tag.Key{KeyMethod, KeyPath, KeyStatus}}

	LineCountView = &view.View{
		Measure:     MetricApiStatus,
		Description: "The number of lines from standard input",
		Aggregation: view.Count(),
		TagKeys:     []tag.Key{KeyMethod, KeyPath, KeyStatus}}
)

func (m StandardMetrics) InitializeMetrics() {

	// Register the views
	if err := view.Register(LatencyView, LineCountView); err != nil {
		log.Fatalf("Failed to register views: %v", err)
	}

	pe, err := prometheus.NewExporter(prometheus.Options{
		Namespace:   m.Namespace,
		ConstLabels: m.ConstLabels,
	})
	if err != nil {
		stderr.Printf("can't start metric exporter!")
	}

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", pe)
		if err := http.ListenAndServe(":8888", mux); err != nil {
			log.Fatalf("Failed to run Prometheus scrape endpoint: %v", err)
		}
	}()

	// 1. Configure exporter to export traces to Zipkin.
	localEndpoint, err := openzipkin.NewEndpoint("go-quickstart", "192.168.1.5:5454")
	if err != nil {
		log.Fatalf("Failed to create the local zipkinEndpoint: %v", err)
	}
	reporter := zipkinHTTP.NewReporter("http://localhost:9411/api/v2/spans")
	ze := zipkin.NewExporter(reporter, localEndpoint)
	trace.RegisterExporter(ze)

	// 2. Configure 100% sample rate, otherwise, few traces will be sampled.
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

}

func sinceInMilliseconds(startTime time.Time) float64 {
	return float64(time.Since(startTime).Nanoseconds()) / 1e6
}

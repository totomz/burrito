package rea

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var mPing = sync.OnceValue[metric.Int64Counter](func() metric.Int64Counter {
	meter := otel.Meter("rea")
	m, err := meter.Int64Counter(
		"ping_sum",
		metric.WithDescription("A test metric"),
	)
	if err != nil {
		panic(err)
	}

	return m
})

func trackPing(ctx context.Context, alabel string) {
	mPing().Add(ctx, 1, metric.WithAttributes(
		attribute.KeyValue{Key: "alabel", Value: attribute.StringValue(alabel)},
	))
}

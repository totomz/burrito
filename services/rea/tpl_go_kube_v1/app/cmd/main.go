package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/totomz/burrito/v2/common"
	"github.com/totomz/burrito/v2/services/rea"
	"github.com/totomz/burrito/v2/telemetry"
)

func main() {
	common.InitConfig("[[.ServiceName]]")

	// Cancel ctx on SIGTERM/SIGINT
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	telemetry.StartOpenTelemetryPrometheus("[[.ServiceName]]")

	service := rea.NewService()

	// Run the service
	go service.Run(ctx)

	// Block until a signal cancels the context
	<-ctx.Done()
	slog.Info("system shutdown")

	// Use a fresh, bounded context for shutdown: ctx is already cancelled here
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer stopCancel()
	service.Stop(stopCtx)
}

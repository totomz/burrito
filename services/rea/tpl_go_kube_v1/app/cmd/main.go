package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/totomz/burrito/common"
	"github.com/totomz/burrito/services/rea"
)

func main() {
	common.InitConfig("[[.ServiceName]]")
	ctx := context.Background()
	common.StartOpenTelemetryPrometheus("[[.ServiceName]]")

	service := rea.NewService()

	// Defer the stop function
	defer service.Stop(ctx)

	// Run the service
	go service.Run(ctx)

	// env := common.GetEnvironment()
	// if env != common.EnvironmentLocal {
	// 	temporalAPIKey := common.MustGetString("temporal.token")
	//
	// 	temporalClient, err := common.GetTemporalClient(ctx, common.GetTemporalNamespace(), env, temporalAPIKey)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	defer temporalClient.Close()
	//
	// 	// Register here workflows and activities
	// 	// meWorker := worker.New(temporalClient, common.TemporalTaskQueuePacioli, worker.Options{})
	// 	// meWorker.RegisterActivityWithOptions(pacioli.UserEventActivityPacioli, activity.RegisterOptions{Name: "UserEventActivityPacioli"})
	// 	// slog.Info("temporal activity registered", "activity", "UserEventActivityPacioli", "queue", common.TemporalTaskQueuePacioli, "namespace", common.GetTemporalNamespace())
	// 	//
	// 	// err = meWorker.Start()
	// 	// if err != nil {
	// 	// 	panic(err)
	// 	// }
	// 	// defer meWorker.Stop()
	// }
	//
	// if env == common.EnvironmentLocal {
	// 	slog.Info("#### LOCAL ENV DETECTED - TEMPORAL DISABLED ####")
	// }

	interupt := make(chan os.Signal, 1)
	signal.Notify(interupt, syscall.SIGTERM, syscall.SIGINT)
	<-interupt

	slog.Info("system shutdown")

}

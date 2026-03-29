package common

import (
	"context"
	"crypto/tls"
	"fmt"
	"go.temporal.io/sdk/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log/slog"
)

/*
	From the Temporal doc:
	Any Worker can pick up any Task on a given Task Queue.
	You must ensure that if a Worker accepts a Task that it can process that task
	using one of its registered Workflows, Activities, or Nexus Operation handlers.
	This means that all Workers listening to a Task Queue must register all Workflows, Activities,
	and Nexus Operations that live on that Queue.

	So:
		- We have a queue for each service;
		- we should have only 2 namespace `hack` and `prod`
*/

const (
	TemporalTaskQueuePacioli  = "pacioli-20250730"
	TemporalTaskQueueKfc      = "kfc-20250730"
	TemporalTaskQueueMolinari = "molinari-20251215"
)

func GetTemporalNamespace() string {
	return fmt.Sprintf("%s.zho7a", GetEnvironment())
}

func GetTemporalClient(_ context.Context, namespace string, env Environment, apiKeyTemporal string) (client.Client, error) {
	log := slog.With("namespace", namespace)

	// option for local env
	// `brew install temporal`
	// `temporal server start-dev`
	if env == EnvironmentLocal && !IsCICD() {
		println("####################################")
		println("# TEMPORAL LOCAL ENV DETECTED		#")
		println("# Start the local temporal client	#")
		println("# `temporal server start-dev`		#")
		println("####################################")
		return client.Dial(client.Options{Logger: log})
	}

	if env == EnvironmentLocal {
		env = EnvironmentHack
	}

	temporalOption := client.Options{
		HostPort:  "eu-west-1.aws.api.temporal.io:7233",
		Namespace: namespace,
		Logger:    log,
		ConnectionOptions: client.ConnectionOptions{
			TLS: &tls.Config{},

			DialOptions: []grpc.DialOption{
				grpc.WithUnaryInterceptor(
					func(ctx context.Context, method string, req any, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
						return invoker(
							metadata.AppendToOutgoingContext(ctx, "temporal-namespace", namespace),
							method,
							req,
							reply,
							cc,
							opts...,
						)
					},
				),
			},
		},
		Credentials: client.NewAPIKeyStaticCredentials(apiKeyTemporal),
	}

	return client.Dial(temporalOption)
}

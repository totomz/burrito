package rea

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/totomz/burrito/common"
	"github.com/totomz/burrito/common/httpserver"
)

var (
	BuildTime string
	GitCommit string
	GitBranch string
)

type Service struct {
	endpoints   map[string]map[string]httpserver.Endpoint
	wsEndpoints map[string]httpserver.WsEndpoint
	Environment string
}

func NewService() *Service {

	noauthChain := []httpserver.InFilter{
		httpserver.InitRequest,
	}

	defaultTokenHandler := []httpserver.InFilter{
		httpserver.InitRequest,
		httpserver.GlobalAuthHandler("", "", getGcloudProjectId()),
	}

	return &Service{
		endpoints: map[string]map[string]httpserver.Endpoint{
			"/public/healthz": {
				"GET": {
					Handler:      handleHealthz,
					InputFilters: noauthChain,
					OutFilters:   httpserver.DefaultOutChain,
				},
			},
			"/healthz": {
				"GET": {
					Handler:      handleHealthz,
					InputFilters: defaultTokenHandler,
					OutFilters:   httpserver.DefaultOutChain,
				},
			},
		},
		wsEndpoints: map[string]httpserver.WsEndpoint{},
		Environment: string(common.GetEnvironment()),
	}
}

// Run is the main function of the service
func (s *Service) Run(ctx context.Context) {
	server := httpserver.NewHttpServer(s)
	port := GetBindPort()

	slog.InfoContext(ctx, "###########################")
	slog.InfoContext(ctx, fmt.Sprintf("Git Branch: %s", GitBranch))
	slog.InfoContext(ctx, fmt.Sprintf("Build Time: %s", BuildTime))
	slog.InfoContext(ctx, fmt.Sprintf("Git Commit: %s", GitCommit))
	slog.InfoContext(ctx, fmt.Sprintf("Listening on 0.0.0.0:%v", port))
	slog.InfoContext(ctx, "###########################")

	server.StartAsync("0.0.0.0", port)
	slog.Info("server started")

	// Wait forever
	<-server.StopChan
	slog.Info("bye")

}

func (s *Service) StartupProbe(_ context.Context, _ *http.Request) (interface{}, error) {
	return "ok", nil
}

func (s *Service) LivenessProbe(ctx context.Context, r *http.Request) (interface{}, error) {
	return s.StartupProbe(ctx, r)
}

func (s *Service) ReadinessProbe(_ context.Context, _ *http.Request) (interface{}, error) {
	return "ok", nil
}

func (s *Service) Endpoints() map[string]map[string]httpserver.Endpoint {
	return s.endpoints
}

func (s *Service) WsEndpoints() map[string]httpserver.WsEndpoint {
	return s.wsEndpoints
}

func (s *Service) DeploymentEnvironment() string {
	return string(common.GetEnvironment())
}

// Stop is the function that stops the service
func (s *Service) Stop(_ context.Context) {
	// nothing to do
}

func handleHealthz(ctx context.Context, _ *http.Request) (interface{}, error) {
	trackPing(ctx, "pong")
	return "ok", nil
}

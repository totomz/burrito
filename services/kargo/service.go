package kargo

import (
	"context"
	"github.com/totomz/template-burrito/common/httpserver"
	"log"
	"net/http"
	"os"
)

var (
	stdout = log.New(os.Stdout, "", log.Lshortfile|log.Ltime)
	stderr = log.New(os.Stderr, "[ERROR]", log.Lshortfile|log.Ltime)
)

type Service struct {
	endpoints map[string]map[string]httpserver.Endpoint
}

func NewService() *Service {
	return &Service{
		endpoints: map[string]map[string]httpserver.Endpoint{},
	}
}

func (s *Service) StartupProbe(_ context.Context, _ *http.Request) (interface{}, error) {
	return "startup", nil
}

func (s *Service) LivenessProbe(_ context.Context, _ *http.Request) (interface{}, error) {
	return "liveness", nil
}
func (s *Service) Endpoints() map[string]map[string]httpserver.Endpoint {
	return s.endpoints
}

func (s *Service) Stdout() *log.Logger {
	return stdout
}

func (s *Service) Stderr() *log.Logger {
	return stderr
}

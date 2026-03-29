package httpserver

import (
	"context"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
)

type MockedService struct {
	endpoints map[string]map[string]Endpoint
}

func (s *MockedService) ReadinessProbe(ctx context.Context, r *http.Request) (interface{}, error) {
	return "readiness", nil
}

func (s *MockedService) WsEndpoints() map[string]WsEndpoint {
	return make(map[string]WsEndpoint)
}
func (s *MockedService) StartupProbe(_ context.Context, _ *http.Request) (interface{}, error) {
	return "startup", nil
}
func (s *MockedService) LivenessProbe(_ context.Context, _ *http.Request) (interface{}, error) {
	return "liveness", nil
}
func (s *MockedService) Endpoints() map[string]map[string]Endpoint {
	return s.endpoints
}
func (s *MockedService) DeploymentEnvironment() string { return "local" }

func TestServerStartStop(t *testing.T) {
	t.Parallel()

	stdout = log.New(os.Stdout, "", log.Lshortfile|log.Ltime)
	stderr = log.New(os.Stdout, "[ERROR]", log.Lshortfile|log.Ltime)
	infilters := []InFilter{
		func(ctx context.Context, request *http.Request) (context.Context, error) {
			return ctx, nil
		},
	}

	outFilters := []OutFilter{
		func(ctx context.Context, request interface{}, err error) ([]byte, int) {
			return []byte(request.(string)), 200
		},
	}

	endpoints := map[string]map[string]Endpoint{
		"/ping": {
			"GET": {
				InputFilters: infilters,
				Handler:      HelloWorld("get"),
				OutFilters:   outFilters,
			},
			"POST": {
				InputFilters: infilters,
				Handler:      HelloWorld("post"),
				OutFilters:   outFilters,
			},
		},
	}

	service := MockedService{endpoints: endpoints}

	// Override logs to not log on files
	server := NewHttpServer(&service)
	server.StartAsync("localhost", 8022)

	err := doGet(t, "http://localhost:8022/ping", nil, 200, []byte("get-hello, world!"))
	if err != nil {
		t.Fatal(err)
	}

	err = doGet(t, "http://localhost:8022/_probe/liveness", nil, 200, []byte("\"liveness\""))
	if err != nil {
		t.Fatal(err)
	}

	err = doGet(t, "http://localhost:8022/_probe/startup", nil, 200, []byte("\"startup\""))
	if err != nil {
		t.Fatal(err)
	}

	err = doPost(t, "http://localhost:8022/ping", 200, []byte("post-hello, world!"))
	if err != nil {
		t.Fatal(err)
	}

	err = server.Stop()
	if err != nil {
		t.Error(err)
	}
}

func doGet(t *testing.T, uri string, authToken *string, expectedStatusCode int, expectedResult []byte) error {

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		t.Fatal(err)
	}

	if authToken != nil {
		req.Header.Set("Authorization", fmt.Sprintf("bearer %s", *authToken))
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if expectedStatusCode != resp.StatusCode {
		return fmt.Errorf("statusCode mismatch: expected %v got: %v", expectedStatusCode, resp.StatusCode)
	}

	if strings.Compare(string(expectedResult), string(data)) != 0 {
		return fmt.Errorf("body mismatch: \nexpected: %v\ngot: %v", string(expectedResult), string(data))
	}
	return nil
}

func doPost(t *testing.T, uri string, expectedStatusCode int, expectedResult []byte) error {
	resp, err := http.Post(uri, "application/json", nil)
	if err != nil {
		return err
	}

	if expectedStatusCode != resp.StatusCode {
		e := fmt.Errorf("statusCode mismatch: expected %v got: %v", expectedStatusCode, resp.StatusCode)
		t.Error(e)
		return e
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if strings.Compare(string(expectedResult), string(data)) != 0 {
		return fmt.Errorf("statusCode mismatch: \nexpected %v\ngot: %v", expectedResult, data)
	}

	return nil
}

func HelloWorld(prefix string) Handler {
	return func(ctx context.Context, request *http.Request) (interface{}, error) {
		return fmt.Sprintf("%s-hello, world!", prefix), nil
	}
}

func TestGetIdToken(t *testing.T) {
	t.Parallel()

	// Get bearer ctoken
	h1 := http.Header{}
	h1.Set("authorization", "Bearer 123123123123123123123")

	if diff := cmp.Diff(GetIdToken(h1), "123123123123123123123"); len(diff) > 0 {
		t.Fatalf("\ntoken not found\n%s", diff)
	}

	// Actually, it gets **only** Bearer Authorization token longer than 10 chars
	h2 := http.Header{}
	h2.Set("Authorization", "12345678901")

	if diff := cmp.Diff(GetIdToken(h2), ""); len(diff) > 0 {
		t.Fatalf("\ntoken not found\n%s", diff)
	}

	// It handles websocket, too (no bearer because header value can't have spaces)
	h3 := http.Header{}
	h3.Set("Sec-WebSocket-Protocol", "authorization, 12345678901, pippo, pasticcio")

	if diff := cmp.Diff(GetIdToken(h3), "12345678901"); len(diff) > 0 {
		t.Fatalf("\ntoken not found\n%s", diff)
	}

	// but it must be the firrst header
	h4 := http.Header{}
	h4.Set("Sec-WebSocket-Protocol", "pippo, authorization, Bearer 12345678901")

	if diff := cmp.Diff(GetIdToken(h4), ""); len(diff) > 0 {
		t.Fatalf("\ntoken not found\n%s", diff)
	}
}

package httpserver

import (
	"context"
	"fmt"
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

func (s *MockedService) StartupProbe(_ context.Context, _ *http.Request) (interface{}, error) {
	return "startup", nil
}

func (s *MockedService) LivenessProbe(_ context.Context, _ *http.Request) (interface{}, error) {
	return "liveness", nil
}
func (s *MockedService) Endpoints() map[string]map[string]Endpoint {
	return s.endpoints
}

func TestServerStartStop(t *testing.T) {

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
	server := NewHttpServer(&service, stdout, stderr)
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

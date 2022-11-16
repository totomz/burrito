package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// InFilter functions are executed before the handler. The output of any last non-nil
type InFilter func(ctx context.Context, request *http.Request) (context.Context, error)

// Handler is responsible to handle a request
type Handler func(ctx context.Context, request *http.Request) (interface{}, error)

// OutFilter functions are processed in order. If
type OutFilter func(ctx context.Context, request interface{}, err error) ([]byte, int)

const (
	CtxTrace = iota
	CtxApiCall
	CtxApiStart
)

type Endpoint struct {
	Handler      Handler
	ErrorHandler Handler
	InputFilters []InFilter
	OutFilters   []OutFilter
}

type HttpServer struct {
	mux      *http.ServeMux
	server   *http.Server
	StopChan chan bool
}

var (
	stdout = log.New(os.Stdout, "", log.Lshortfile)
	stderr = log.New(os.Stderr, "[error]", log.Lshortfile)
)

type Service interface {
	StartupProbe(ctx context.Context, r *http.Request) (interface{}, error)
	LivenessProbe(ctx context.Context, r *http.Request) (interface{}, error)
	Endpoints() map[string]map[string]Endpoint
}

// func (s *Service) StartupProbe(_ context.Context, _ *http.Request) (interface{}, error) {
//	return "ping", nil
// }
//
// func (s *Service) LivenessProbe(_ context.Context, _ *http.Request) (interface{}, error) {
//	return "pong", nil
// }
//
// func (s *Service) Endpoints() map[string]map[string]Endpoint {
//	return map[string]map[string]Endpoint{}
// }

func InitRequest(ctx context.Context, request *http.Request) (context.Context, error) {

	apiCall := request.Method + " " + request.URL.Path
	trace := time.Now().Unix()

	newctx := context.WithValue(ctx, CtxTrace, trace)
	newctx = context.WithValue(newctx, CtxApiCall, apiCall)
	newctx = context.WithValue(newctx, CtxApiStart, time.Now())

	return newctx, nil
}

func Jsonize(_ context.Context, request interface{}, err error) ([]byte, int) {

	var data []byte
	httpStatuscode := http.StatusOK

	if err != nil {

		switch err.(type) {
		case ErrUnauthorized:
			httpStatuscode = (err.(ErrUnauthorized)).HttpCode()
			data, _ = json.Marshal(err)
		case ErrBadRequest:
			httpStatuscode = (err.(ErrBadRequest)).HttpCode()
			data, _ = json.Marshal(err)
		case ErrNotFound:
			httpStatuscode = (err.(ErrNotFound)).HttpCode()
			data, _ = json.Marshal(err)

		default:
			httpStatuscode = http.StatusInternalServerError
			stderr.Printf("request error - %v ", err)
			data = []byte(fmt.Sprintf(`{"message": "%s"}`, err.Error()))
		}
	} else {
		data, err = json.Marshal(request)
		if err != nil {
			panic(err) // boh, someone will fix this
		}
	}

	return data, httpStatuscode
}

// func LogResponseStats(ctx context.Context, _ interface{}, _ error) ([]byte, int) {
// 	now := time.Now()
// 	elapsed := time.Duration(0)
//
// 	trace := ctx.Value(CtxTrace).(int64)
// 	api := ctx.Value(CtxApiCall).(string)
//
// 	start, hasStart := ctx.Value(CtxApiStart).(time.Time)
// 	if hasStart {
// 		elapsed = now.Sub(start)
// 	}
//
// 	stdout.Printf("    :request %s %v [%v]", api, elapsed, trace)
//
// 	return nil, 0
// }

var DefaultOutChain = []OutFilter{Jsonize /*, LogResponseStats*/}

func NewHttpServer(service Service) *HttpServer {

	// fileServer := http.FileServer(http.Dir("./webmin/dist/"))
	mux := http.NewServeMux()

	// The notFoundHandler logs requests for unmapped urls
	// We should move the dashboard somewhere else, or under a different path.
	mux.HandleFunc("/", notFoundHandler)
	// mux.Handle("/admin", http.StripPrefix("/", fileServer))
	// mux.Handle("/admin", http.StripPrefix("/admin", fileServer))
	// mux.Handle("/", http.StripPrefix("/", fileServer))

	httpEndpoints := service.Endpoints()

	// Force Liveness and Startup endpoints
	httpEndpoints["/_probe/liveness"] = map[string]Endpoint{
		"GET": {
			InputFilters: []InFilter{InitRequest},
			Handler:      service.LivenessProbe,
			OutFilters:   DefaultOutChain,
		},
	}
	httpEndpoints["/_probe/startup"] = map[string]Endpoint{
		"GET": {
			InputFilters: []InFilter{InitRequest},
			Handler:      service.StartupProbe,
			OutFilters:   DefaultOutChain,
		},
	}

	for uri := range httpEndpoints {
		stdout.Printf("registering endpoint %s", uri)
		endpoints := httpEndpoints[uri]
		mux.HandleFunc(uri, func(w http.ResponseWriter, r *http.Request) {

			// ctx is the Context for this handler. Calling cancel closes the
			// ctx.Done channel, which is the cancellation signal for requests
			// started by this handler.
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel() // Cancel ctx as soon as handleSearch returns.

			var requestInError error
			var handlerResult interface{}

			setupCORS(&w, r)
			if (*r).Method == "OPTIONS" {
				return
			}

			// Get the handler mapped to the requested HTTP method
			handler, endpointFound := endpoints[strings.ToUpper(r.Method)]
			if !endpointFound {
				requestInError = fmt.Errorf("method %s not found", strings.ToUpper(r.Method))
				stderr.Printf("method %s not found for endpoint %s", r.Method, r.RequestURI)
			}

			if requestInError == nil {
				for _, filter := range handler.InputFilters {
					ctx, requestInError = filter(ctx, r)
					if requestInError != nil {
						break
					}
				}
			}

			if requestInError == nil {
				handlerResult, requestInError = handler.Handler(ctx, r)
			}

			var responseData []byte
			responseStatus := 200

			for _, outFilter := range handler.OutFilters {
				res, stat := outFilter(ctx, handlerResult, requestInError)
				if res != nil {
					responseData = res
				}
				if stat != 0 {
					responseStatus = stat
				}
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(responseStatus)
			_, err := w.Write(responseData)
			if err != nil {
				stderr.Printf("panic error? %v", err)
			}
		})
	}

	return &HttpServer{
		mux: mux,
	}
}

func setupCORS(w *http.ResponseWriter, _ *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func (s *HttpServer) StartAsync(host string, port int) {
	server := &http.Server{Addr: fmt.Sprintf("%s:%v", host, port), Handler: s.mux}
	wg := sync.WaitGroup{}
	wg.Add(1)
	if os.Getenv("BIND_SSL") == "ssl" {
		go func() {
			stdout.Println("starting SSL mode")
			// always returns error. ErrServerClosed on graceful close
			dir, _ := os.Getwd()
			path := strings.Join([]string{dir, "certs"}, string(filepath.Separator))

			// Go up on level if there are no certs
			if _, err := os.Stat(path); os.IsNotExist(err) {
				stdout.Printf("TLS certs not found 0/1 in path %s", path)
				path = strings.Join([]string{dir, "..", "certs"}, string(filepath.Separator))
				if _, err = os.Stat(path); os.IsNotExist(err) {
					stderr.Printf("TLS certs not found 1/1 in path %s - aborting", path)
					panic("no TLS certs found")
				}
			}

			certFile := strings.Join([]string{path, "live", "bbk.my-ideas.it", "fullchain.pem"}, string(filepath.Separator))
			keyFile := strings.Join([]string{path, "live", "bbk.my-ideas.it", "privkey.pem"}, string(filepath.Separator))

			wg.Done() // Call Done() here because Serve() is blocking. Done() will free the main thread
			if err := server.ListenAndServeTLS(certFile, keyFile); err != http.ErrServerClosed {
				// unexpected error. port in use?
				stderr.Printf("can't start the service: %v", err)
			}
			stdout.Printf("Server started on %s:%v", host, port)

		}()
	} else {
		go func() {
			wg.Done() // Call Done() here because Serve() is blocking. Done() will free the main thread
			if err := server.ListenAndServe(); err != http.ErrServerClosed {
				stderr.Printf("can't start the service: %v", err)
			}
			stdout.Printf("Server started on %s:%v", host, port)
		}()
	}

	wg.Wait()
	time.Sleep(1 * time.Second)
	// Setting up signal capturing
	stop := make(chan bool, 1)
	// signal.Notify(stop, os.Interrupt)

	s.server = server
	s.StopChan = stop
}

func (s *HttpServer) Stop() error {
	err := s.server.Shutdown(context.Background())
	if err != nil {
		stderr.Printf("error stopping the server: %v", err)
	}
	s.StopChan <- true
	return nil
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	status := http.StatusNotFound
	w.WriteHeader(status)
	stdout.Printf("NotFoundHandler: %s %s Not Found", r.Method, r.RequestURI)
	_, _ = fmt.Fprint(w, "not found")
}

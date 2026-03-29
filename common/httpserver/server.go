package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/totomz/burrito/common"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// InFilter functions are executed before the handler. The output of any last non-nil
type InFilter func(ctx context.Context, request *http.Request) (context.Context, error)

// Handler deprecated is responsible to handle an http request.
// DO NOT USE DIRECTLY! Use WrapHandler(HandlerFunc) to have typed return values
type Handler func(ctx context.Context, request *http.Request) (interface{}, error)

// HandlerFunc is responsible to handle an http request
type HandlerFunc[T any] func(ctx context.Context, request *http.Request) (T, error)

func WrapHandler[T any](handler HandlerFunc[T]) Handler {
	return func(ctx context.Context, req *http.Request) (interface{}, error) {
		result, err := handler(ctx, req)
		return result, err
	}
}

// WsHandler is a low level handler for websockets
type WsHandler func(ctx context.Context, w http.ResponseWriter, r *http.Request)

// OutFilter functions are processed in order. If
type OutFilter func(ctx context.Context, request interface{}, err error) ([]byte, int)

const (
	CtxApiStart    = "__CTX_APISTART"
	CtxRequestId   = "__CTX_REQUEST_ID"
	CtxDeviceId    = "__CTX_DEVICE_ID"
	CtxRequestUser = "__CTX_REQUEST_USER"
	CtxRequestLog  = "__CTX_REQUEST_LOGLINE"
	CtxRequestTags = "__CTX_REQUEST_TAGS"
)

type Endpoint struct {
	Handler      Handler
	ErrorHandler Handler
	InputFilters []InFilter
	OutFilters   []OutFilter
}

type WsEndpoint struct {
	WsHandler    WsHandler
	InputFilters []InFilter

	// ErrorHandler Handler
	// InputFilters []InFilter
	// OutFilters   []OutFilter
}

type HttpServer struct {
	mux      *mux.Router
	server   *http.Server
	StopChan chan bool
}

var (
	stdout = log.New(os.Stdout, "", log.Lshortfile)
	stderr = log.New(os.Stderr, "[error] ", log.Lshortfile)
)

type Service interface {
	StartupProbe(ctx context.Context, r *http.Request) (interface{}, error)
	LivenessProbe(ctx context.Context, r *http.Request) (interface{}, error)
	ReadinessProbe(ctx context.Context, r *http.Request) (interface{}, error)
	Endpoints() map[string]map[string]Endpoint
	WsEndpoints() map[string]WsEndpoint
	DeploymentEnvironment() string
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
	newctx := context.WithValue(ctx, CtxApiStart, time.Now())

	reqId := request.Header.Get("X-Request-ID")
	if reqId == "" {
		uid, err := uuid.NewRandom()
		if err == nil {
			reqId = uid.String()
		}
	}

	newctx = context.WithValue(newctx, CtxRequestId, reqId)

	ua := request.Header.Get("User-Agent")
	deviceId := request.Header.Get("X-Device-ID")

	requestLog := fmt.Sprintf(`http_method="%s" url="%s" hostname="%s" agent="%s" device_id="%s" req_id="%s" origin="%s"`, request.Method, request.URL.RequestURI(), request.Host, ua, deviceId, reqId, request.RemoteAddr)
	metricTags := []attribute.KeyValue{
		semconv.HTTPMethod(request.Method),
		semconv.URLPath(request.URL.Path),
		semconv.UserAgentOriginal(ua),
	}

	deploymentEnvironment := ctx.Value(semconv.DeploymentEnvironmentKey)
	if deploymentEnvironment != nil {
		metricTags = append(metricTags, semconv.DeploymentEnvironment(deploymentEnvironment.(string)))
	}

	newctx = context.WithValue(newctx, CtxDeviceId, deviceId)
	newctx = context.WithValue(newctx, CtxRequestLog, requestLog)
	newctx = context.WithValue(newctx, CtxRequestTags, metricTags)

	return newctx, nil
}

func Jsonize(_ context.Context, request interface{}, err error) ([]byte, int) {

	var data []byte
	httpStatuscode := http.StatusOK

	if err == nil {
		data, err = json.Marshal(request)
		if err != nil {
			stderr.Printf("can't marshal response to json - %s", err.Error())
			data = []byte("")
		}
		return data, httpStatuscode
	}

	httpStatuscode = errorToHttpStatuscode(err)
	data, _ = json.Marshal(err)
	if httpStatuscode == http.StatusInternalServerError {
		stderr.Printf("request error - %v ", err)
		data = []byte(fmt.Sprintf(`{"message": "%s"}`, err.Error()))
	}

	return data, httpStatuscode
}

func LogResponseStats(ctx context.Context, _ interface{}, err error) ([]byte, int) {

	httpStatuscode := http.StatusOK
	if err != nil {
		httpStatuscode = errorToHttpStatuscode(err)
	}

	start, hasStart := ctx.Value(CtxApiStart).(time.Time)
	elapsedMs := 0.0
	if hasStart {
		elapsedMs = sinceInMilliseconds(start)
		// stats.Record(ctx, MetricApiLatency.M(elapsedMs))
	}

	requestLogLine := ctx.Value(CtxRequestLog).(string)
	if common.IsKube() || !strings.Contains(requestLogLine, "/conversations/messages") {
		println(fmt.Sprintf(`level=INFO source=server.go:164: status=%v %s time_ms="%v"`, httpStatuscode, requestLogLine, elapsedMs))
	}

	var tags []attribute.KeyValue
	ctxtags := ctx.Value(CtxRequestTags)
	if ctxtags != nil {
		tags = ctxtags.([]attribute.KeyValue)
	}

	tags = append(tags, semconv.HTTPStatusCode(httpStatuscode))

	mHttpServerDurationS.Record(ctx, elapsedMs/1000, metric.WithAttributes(tags...))

	return nil, 0
}

var DefaultOutChain = []OutFilter{Jsonize, LogResponseStats}

func NewHttpServer(service Service /*, metrics *StandardMetrics*/) *HttpServer {

	initMetrics()

	// fileServer := http.FileServer(http.Dir("./webmin/dist/"))
	// mux := http.NewServeMux()
	r := mux.NewRouter()

	wdir, _ := os.Getwd()
	staticContents := fmt.Sprintf("%s%sweb_static", wdir, string(os.PathSeparator))

	s, err := os.Stat(staticContents)
	if err == nil && s.IsDir() {
		stdout.Printf("serving static contents from: %s", staticContents)
		r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(staticContents))))
	}

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
	httpEndpoints["/_probe/readiness"] = map[string]Endpoint{
		"GET": {
			InputFilters: []InFilter{InitRequest},
			Handler:      service.ReadinessProbe,
			OutFilters:   DefaultOutChain,
		},
	}

	for uri := range httpEndpoints {
		stdout.Printf("registering endpoint %s", uri)
		endpoints := httpEndpoints[uri]
		r.HandleFunc(uri, func(w http.ResponseWriter, r *http.Request) {

			// r.Context() is the context bound to the http request.
			// When the client close the connection, either by mistake or by an IP error, r.Context() is closed.
			// But we pass this context down also to async requests, eg to firebase stuff, and at the moment the
			// closed context is never handled.
			// So, here we create a new context, not bounded to the request.
			// If you reach this comment, and we have just done the Serie A round,
			// feel free to pick and fix how we handle the context
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

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

			select {
			case <-r.Context().Done():
				slog.InfoContext(ctx, "Initial request context is Done, skipping response to the client.")
				return
			default:
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(responseStatus)
				_, errWriteResponse := w.Write(responseData)
				if errWriteResponse != nil {
					stderr.Printf("error sending reply to the client: %v", errWriteResponse)
				}
			}
		})
	}

	wsEndpoints := service.WsEndpoints()
	for uri := range wsEndpoints {
		handler := wsEndpoints[uri]
		r.HandleFunc(uri, func(w http.ResponseWriter, r *http.Request) {

			setupCORS(&w, r)
			if (*r).Method == "OPTIONS" {
				return
			}

			ctx := context.WithValue(context.Background(), semconv.DeploymentEnvironmentKey, service.DeploymentEnvironment())

			var requestInError error

			for _, filter := range handler.InputFilters {
				ctx, requestInError = filter(ctx, r)
				if requestInError != nil {

					var redirect ErrRedirect
					isRedirect := errors.As(requestInError, &redirect)

					if isRedirect {
						stdout.Printf("request redirected to : %s", redirect.RedirectUrl)
						http.Redirect(w, r, redirect.RedirectUrl, http.StatusTemporaryRedirect)
						return
					}

					stdout.Printf("request blocked by initfilter: %v", requestInError)
					body, status := Jsonize(ctx, nil, requestInError)
					w.WriteHeader(status)
					_, _ = w.Write(body)
					return
				}
			}

			handler.WsHandler(ctx, w, r)
		})

	}

	// nofFoundHandler := func(w http.ResponseWriter, r *http.Request) {}
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stdout.Printf(`msg="not found" method="%s" uri="%s" source="%s"`, r.Method, r.RequestURI, r.RemoteAddr)
		w.WriteHeader(http.StatusNotFound)
	})

	return &HttpServer{
		mux: r,
	}
}

func setupCORS(w *http.ResponseWriter, _ *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Request-Id, X-Device-ID, user-agent, ngrok-skip-browser-warning, x-happ")
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
				panic(fmt.Sprintf("can't start the service: %v", err))
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

// func notFoundHandler(w http.ResponseWriter, r *http.Request) {
// 	status := http.StatusNotFound
// 	w.WriteHeader(status)
// 	stdout.Sprintf("NotFoundHandler: %s %s Not Found", r.Method, r.RequestURI)
// 	_, _ = fmt.Fprint(w, "not found")
// }

func GetIdToken(headers http.Header) string {

	idToken := headers.Get("Authorization")
	if idToken != "" || len(idToken) > 10 {
		// Regular bearer token!
		// clean it up and return it
		splitToken := strings.Split(idToken, " ")
		if len(splitToken) != 2 || !strings.EqualFold(splitToken[0], "Bearer") {
			return ""
		}
		return strings.TrimSpace(splitToken[1])
	}

	// TL;DR Javascript allows us to specify only the `Sec-WebSocket-Protocol`, we exploit it adding the idToken
	// There is no method in the JavaScript WebSockets API for specifying additional headers for the client/browser to send.
	// Only he HTTP path ("GET /xyz") and protocol header ("Sec-WebSocket-Protocol") can be specified in the WebSocket constructor.
	// The `Sec-WebSocket-Protocol` header is defined as `var ws = new WebSocket("ws://example.com/path", ["Authorization", "Bearer XXXX"]);`
	// That results in
	// `Sec-WebSocket-Protocol: Authorization, Bearer XXXX`
	// that sometimes is extended to be used in websocket specific authentication, at least by Deepgram and by us.
	//
	// Resta di stucco, è un barbatrucco!
	wsIdToken := headers.Get("Sec-WebSocket-Protocol")
	if wsIdToken == "" {
		return ""
	}

	parts := strings.Split(wsIdToken, ",")
	if len(parts) < 2 || strings.ToLower(parts[0]) != "authorization" {
		return ""
	}

	idToken = strings.TrimSpace(parts[1])

	return idToken
}

// func GetUserFromContext(ctx context.Context) (talktome.User, error) {
// 	u := ctx.Value(CtxRequestUser)
// 	if u == nil {
// 		return talktome.User{}, errors.New("no user found")
// 	}
//
// 	user, isCastOk := u.(talktome.User)
// 	if !isCastOk {
// 		return talktome.User{}, errors.New("no user found - bad cast")
// 	}
//
// 	return user, nil
// }

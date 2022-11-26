package kargo

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/form3tech-oss/jwt-go"
	"github.com/totomz/template-burrito/common/httpserver"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.opencensus.io/trace"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	stdout = log.New(os.Stdout, "", log.Lshortfile|log.Ltime)
	stderr = log.New(os.Stderr, "[ERROR]", log.Lshortfile|log.Ltime)

	DefaultJwtInChain = []httpserver.InFilter{httpserver.InitRequest, AuthJwt}
	NoAuthInChain     = []httpserver.InFilter{httpserver.InitRequest}

	// Custom metrics
	MetricPong = stats.Int64("pong", "", stats.UnitDimensionless)
	KeyBom, _  = tag.NewKey("KeyBom")
	PongView   = &view.View{
		Name:        "pong",
		Measure:     MetricPong,
		Description: "The counts of Pongs",
		Aggregation: view.Count(),
		TagKeys:     []tag.Key{KeyBom},
	}

	rnd = rand.New(rand.NewSource(time.Now().UnixMilli()))
)

const (
	CtxUserId = "__CtxUser"
)

type Service struct {
	endpoints map[string]map[string]httpserver.Endpoint
}

func NewService() *Service {

	// Register some custom metrics
	// Register the views
	if err := view.Register(PongView); err != nil {
		log.Fatalf("Failed to register views: %v", err)
	}
	return &Service{
		endpoints: map[string]map[string]httpserver.Endpoint{
			"/hello": {
				"GET": {
					Handler:      HelloProtected,
					InputFilters: DefaultJwtInChain,
					OutFilters:   httpserver.DefaultOutChain,
				},
			},
			"/test/ciao": {
				"GET": {
					Handler:      testStatus,
					InputFilters: NoAuthInChain,
					OutFilters:   httpserver.DefaultOutChain,
				},
			},
			"/test/trace": {
				"GET": {
					Handler:      testTrace,
					InputFilters: NoAuthInChain,
					OutFilters:   httpserver.DefaultOutChain,
				},
			},
		},
	}
}

func testStatus(ctx context.Context, request *http.Request) (interface{}, error) {
	// The context inherited the labels set in the InitRequest()N (like the env, the path)
	// Create a new context if you whish to clean the labels
	stats.Record(ctx, MetricPong.M(1))

	coed := request.URL.Query().Get("status")
	switch coed {
	case "200":
		return "200", nil
	case "404":
		return nil, httpserver.ErrNotFound{Message: "Not Found"}
	case "400":
		return nil, httpserver.ErrBadRequest{Message: "Bad Request"}
	default:
		return nil, errors.New("status not mapped")
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

// verifyJWTRSA Verify a JWT token using an RSA public key
func verifyJWTRSA(token string, publicKey string) (bool, *jwt.Token, error) {
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unknown signing method")
		}

		key, err := jwt.ParseRSAPublicKeyFromPEM([]byte(publicKey))
		if err != nil {
			return nil, err
		}

		return key, nil
	})

	if err != nil {
		return false, &jwt.Token{}, err
	}

	if !parsedToken.Valid {
		return false, &jwt.Token{}, errors.New("invalid jwt token")
	}

	return true, parsedToken, nil
}

func AuthJwt(ctx context.Context, request *http.Request) (context.Context, error) {

	if request.Method == "OPTIONS" {
		return ctx, nil
	}

	// remove `bearer ` from the token
	reqToken := request.Header.Get("Authorization")

	if reqToken == "" || len(reqToken) < 10 {
		return ctx, httpserver.ErrUnauthorized{Message: "Missing or invalid Authorization Header"}
	}
	splitToken := strings.Split(reqToken, " ")

	if len(splitToken) != 2 || !strings.EqualFold(splitToken[0], "Bearer") {
		return ctx, httpserver.ErrUnauthorized{Message: "Invalid Authorization Bearer Header"}
	}

	reqToken = splitToken[1]

	// https://syncaltest.eu.auth0.com/.well-known/jwks.json
	const publicKey = `-----BEGIN CERTIFICATE-----
MIIDCTCCAfGgAwIBAgIJQYVFkCZQx7ePMA0GCSqGSIb3DQEBCwUAMCIxIDAeBgNVBAMTF3N5bmNhbHRlc3QuZXUuYXV0aDAuY29tMB4XDTIyMTExMjE1MTk0MVoXDTM2MDcyMTE1MTk0MVowIjEgMB4GA1UEAxMXc3luY2FsdGVzdC5ldS5hdXRoMC5jb20wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQC3PtCy917f7u4lCu9u+MvB1RZ8znwFwKfl5qP+Br9Fez1l96Jb8onN1kx37X7ZZx9tQ09Vwo3mo1W/W4LuTQGeubVLOdXCk05V7aA7scRPSLssP92aOeRPH0dz4CUjbvM/Depih7h1c/x0cXjTMQPyTvfRbED4Y3vLlmTyg6tR4U9hVKumomQtLyNfLNQ56iM47DECzMZ1ojwIyjukmOFrROzdZSEzs0uW0ThpiXng2Zp70oKQOiuWLFIKWgbxw77l3ALz2JdQXL4x7wx17qRYqjOdjJfsMQBQLdetdAE98P/lAC09N3SzZqi04xXaKGX2H2PA518KeyA9WW9IuAnbAgMBAAGjQjBAMA8GA1UdEwEB/wQFMAMBAf8wHQYDVR0OBBYEFMb5Fl2bfbiSglAQ5Iv58EmVvR0TMA4GA1UdDwEB/wQEAwIChDANBgkqhkiG9w0BAQsFAAOCAQEAQgfnV6VM0jKpuzblKAPYzTalzfZJukC4JzCBS+zR2sEHhSrJORcd/KSVptIIWoevmdkih3MU2Gcuz/aZgmGY0mXdRSSMQdiul7D0bzWkeqe4yDwys8iOwlV4LsQEDavkGgv5GLtyxEPHawbHdL1k0Wkc1ev75tFwEctz82nXYXFG4DeOjvg2czXDQ8slrs7fG9YCGCjEIJWfplvP+ibd1XTqemwnhMzoDNqjtJgpAM/vlhnGx4B8hF86awISloL+RMfueTLUM4dqv6SGIoOvnVaQkEgCGQtmt6PoPu7/QeIfIDj8pHZvEi2O9WSmNIqBcqR2YPXEg4PtcgtN+tyNIQ==
-----END CERTIFICATE-----`

	isValid, parsedToken, err := verifyJWTRSA(reqToken, publicKey)
	if err != nil {
		stderr.Printf("error while verifying JWT token - %v", err)
		return ctx, httpserver.ErrUnauthorized{Message: fmt.Sprintf("JWT Token is not valid: %v", err)}
	}

	if !isValid || err != nil {
		return nil, httpserver.ErrUnauthorized{Message: "Token is not invalid"}
	}

	claims := parsedToken.Claims.(jwt.MapClaims)

	newctx := context.WithValue(ctx, CtxUserId, claims["sub"].(string))
	return newctx, nil
}

func HelloProtected(ctx context.Context, _ *http.Request) (interface{}, error) {
	userId := ctx.Value(CtxUserId)
	stdout.Printf("got request from user %s", userId)

	return []string{"blue", "green", "yellow", "cow"}, nil
}

func testTrace(ctx context.Context, _ *http.Request) (interface{}, error) {
	ctx, span := trace.StartSpan(ctx, "main")
	defer span.End()
	workA(ctx)
	return "done", nil
}

func workA(ctx context.Context) {
	ctx, span := trace.StartSpan(ctx, "workA")
	defer span.End()
	time.Sleep(time.Duration(rnd.Intn(3500)) * time.Millisecond)
	workb(ctx)
}

func workb(ctx context.Context) {
	ctx, span := trace.StartSpan(ctx, "workB")
	defer span.End()
	time.Sleep(time.Duration(rnd.Intn(3500)) * time.Millisecond)
	workc(ctx)
}

func workc(ctx context.Context) {
	ctx, span := trace.StartSpan(ctx, "workC")
	defer span.End()
	time.Sleep(time.Duration(rnd.Intn(3500)) * time.Millisecond)
}

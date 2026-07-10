package httpserver

import (
	"context"
	"github.com/golang-jwt/jwt"
	"log/slog"
	"net/http"
)

type UserId string

// GlobalAuthHandler is a middleware that checks for a valid JWT token in the Authorization header.
// Currently it supports Firebase and Auth0 JWT tokens (check auth_auth0.go and auth_firebase.go).
func GlobalAuthHandler(auth0Audience string, auth0CertUrl string, projectId string) InFilter {
	jwtHandler := func(ctx context.Context, request *http.Request) (context.Context, error) {
		// allow OPTIONS requests to allow CORS
		if request.Method == "OPTIONS" {
			return ctx, nil
		}

		idToken := GetIdToken(request.Header)
		if idToken == "" || len(idToken) < 10 {
			slog.Info("missing or invalid authorization token", "authorization_header", idToken)
			return ctx, ErrUnauthorized{Message: "Missing or invalid Authorization Header"}
		}

		// parse the token without validating the signature (to check the issuer)
		token, err := jwt.Parse(idToken, nil)
		if err != nil && err.(*jwt.ValidationError).Errors != jwt.ValidationErrorUnverifiable {
			return ctx, ErrUnauthorized{Message: "Invalid JWT token"}
		}

		// check if the token is a firebase token
		// TODO Dex
		println(token)

		// if it reaches this point, it means neither auth0 or firebase returned a valid auth token
		return ctx, ErrUnauthorized{Message: "No valid JWT token"}
	}
	return jwtHandler

}

package httpserver

import (
	"context"
	"log/slog"
	"net/http"
)

// StaticTokenAuthHandler return a handler that check for a token defined in the input map.
// Only the key is matched
func StaticTokenAuthHandler(allowedTokens map[string]string) InFilter {
	return func(ctx context.Context, request *http.Request) (context.Context, error) {
		// allow cors
		if request.Method == "OPTIONS" {
			return ctx, nil
		}

		idToken := GetIdToken(request.Header)
		if idToken == "" || len(idToken) < 10 {
			slog.Info("missing or invalid authorization token", "authorization_header", idToken)
			return ctx, ErrUnauthorized{Message: "Missing or invalid Authorization Header"}
		}

		serviceAccount, exists := allowedTokens[idToken]
		if exists {
			slog.Info("request authenticated", "service_account", serviceAccount)
			return ctx, nil
		}

		// any other case, ciaone
		return ctx, ErrUnauthorized{Message: "Missing or invalid Authorization Header"}
	}
}

package common

import (
	"context"
	"errors"
	"fmt"
	"github.com/openbao/openbao/api/v2"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func GetSecret(ctx context.Context, secretKey string) (string, error) {
	if strings.HasPrefix(secretKey, "bao:") {
		return GetSecretBao(ctx, secretKey)
	}

	secretEnv := os.Getenv(secretKey)
	if secretEnv == "" {
		return "", errors.New("secret not found")
	}

	return secretEnv, nil
}

// GetSecretBao retrieve a secret from OpenBAO
// secretKey is the secret urn: bao://<mount>/<path>#field
// BAO_ADDRES and BAO_TOKEN are from env
func GetSecretBao(ctx context.Context, secretKey string) (string, error) {

	u, err := url.Parse(secretKey)
	if err != nil {
		return "", fmt.Errorf("invalid secret uri: %w", err)
	}
	if u.Fragment == "" {
		return "", errors.New("missing fragment")
	}

	// BAO_ADDRES and BAO_TOKEN are from env
	cfg := api.DefaultConfig()
	client, err := api.NewClient(cfg)
	if err != nil {
		return "", fmt.Errorf("error connecting to OpenBAO: %w", err)
	}

	if client.Token() == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		t, err := os.ReadFile(filepath.Join(home, ".vault-token"))
		if err != nil {
			return "", fmt.Errorf("no BAO_TOKEN and no ~/.vault-token: %w", err)
		}
		client.SetToken(strings.TrimSpace(string(t)))
	}

	s, err := client.KVv2(u.Host).Get(ctx, strings.TrimPrefix(u.Path, "/"))
	if err != nil {
		return "", fmt.Errorf("error getting the secret: %w", err)
	}

	v, ok := s.Data[u.Fragment].(string)
	if !ok {
		return "", fmt.Errorf("fragment %s not found", u.Fragment)
	}
	return v, nil
}

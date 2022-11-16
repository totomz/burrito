package burrito_common

import (
	"fmt"
	"os"
)

// GetenvOrDefault will return the environmental variable
// by the key, otherwise returns with a default value
func GetenvOrDefault(key, defaultValue string) string {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	return v
}

// GetenvOrDie will return the environmental variable
// by the key, otherwise will panic
func GetenvOrDie(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("missing required env var %s", key))
	}
	return v
}

// GetEnv will return the environmental variable
// by the key if defined, otherwise return an error
func GetEnv(k string) (string, error) {
	v := os.Getenv(k)
	if v == "" {
		return "", fmt.Errorf("%s environment variable not set", k)
	}
	return v, nil
}

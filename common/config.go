package common

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"log/slog"
	"os"
	"path"
	"strconv"
	"strings"
)

// _config holds the whole configuration file parsed as a generic map
var _config map[string]interface{}

// ConfigMiddleware retrieve a secret value from a secret provider (see config/aws)
type ConfigMiddleware func(configKey string, configValue string) (string, error)

type Options struct {
	// Paths define a list of paths (without the file name) to look for the configuration file
	Paths []string

	// FileName define the filename to look for the configuration
	FileName string

	// EnvPrefix if defined, try tto resolv any configuration key with an env with this prefix before
	EnvPrefix string

	// ConfigMiddleware handle secrets (still in development)
	ConfigMiddleware ConfigMiddleware
}

// InitConfig initialize the configuration by reading a yaml file
// The configuration is
func InitConfig(serviceName string) {
	InitConfigWithOptions(Options{
		Paths: []string{
			".",
			path.Join("envs"),
			path.Join("..", "env"),
			path.Join("..", "../", "env"),
			path.Join("/", "etc"),
		},
		FileName:  serviceName + string(GetEnvironment()) + ".yaml",
		EnvPrefix: "",
		ConfigMiddleware: func(_ string, configValue string) (string, error) {
			// default no-op
			return configValue, nil
		},
	})
}

func InitConfigWithOptions(options Options) {
	SetDefaultLogger()

	var data []byte
	var err error

	for _, p := range options.Paths {
		data, err = os.ReadFile(path.Join(p, options.FileName))
		if err != nil {
			// file not found or impossible to read
			continue
		}
		break
	}

	if data == nil {
		slog.Info("configuration file not found", "filename", options.FileName, "paths", options.Paths)
		_config = map[string]interface{}{}
		return
	}

	_config = map[string]interface{}{}
	if err = yaml.Unmarshal(data, &_config); err != nil {
		panic(fmt.Errorf("failed to parse configuration file %s: %w", options.FileName, err))
	}
}

// Get returns (value, true) for the given key, or (zero value, false) if the
// key is missing or cannot be converted to T.
// An environment variable (dots replaced by underscores) always overrides the
// value found in the configuration file.
func Get[T any](key string) (T, bool) {
	var zero T

	// env overrides the config file
	if envValue := os.Getenv(strings.ReplaceAll(key, ".", "_")); envValue != "" {
		return parseValue[T](envValue)
	}

	// fallback: lookup in the parsed config file
	raw, ok := lookup(key)
	if !ok {
		return zero, false
	}
	typed, ok := raw.(T)
	if !ok {
		slog.Error("config value has an unexpected type", "key", key, "value", raw)
		return zero, false
	}
	return typed, true
}

// MustGet panic if the configuration key does not exists
func MustGet[T any](key string) T {
	val, found := Get[T](key)
	if !found {
		panic(fmt.Errorf("configuration key %s not found", key))
	}
	return val
}

// GetOr return the configuration associated to the key, or the default value
func GetOr[T any](key string, defaultValue T) T {
	val, found := Get[T](key)
	if !found {
		return defaultValue
	}
	return val
}

// lookup resolves a dotted key (e.g. "temporal.host") against _config.
func lookup(key string) (interface{}, bool) {
	var current interface{} = _config
	for _, part := range strings.Split(key, ".") {
		m, ok := current.(map[string]interface{})
		if !ok {
			return nil, false
		}
		current, ok = m[part]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

// parseValue converts a raw string (from an env variable) into T.
// Slice types are comma-separated.
func parseValue[T any](s string) (T, bool) {
	var zero T
	var result any
	var err error

	switch any(zero).(type) {
	case string:
		result = s
	case int:
		var v int64
		v, err = strconv.ParseInt(s, 10, 0)
		result = int(v)
	case int32:
		var v int64
		v, err = strconv.ParseInt(s, 10, 32)
		result = int32(v)
	case int64:
		result, err = strconv.ParseInt(s, 10, 64)
	case float32:
		var v float64
		v, err = strconv.ParseFloat(s, 32)
		result = float32(v)
	case float64:
		result, err = strconv.ParseFloat(s, 64)
	case []string:
		result = splitList(s)
	case []int:
		var v []int64
		if v, err = parseIntList(s, 0); err == nil {
			out := make([]int, len(v))
			for i := range v {
				out[i] = int(v[i])
			}
			result = out
		}
	case []int32:
		var v []int64
		if v, err = parseIntList(s, 32); err == nil {
			out := make([]int32, len(v))
			for i := range v {
				out[i] = int32(v[i])
			}
			result = out
		}
	case []int64:
		result, err = parseIntList(s, 64)
	case []float32:
		var v []float64
		if v, err = parseFloatList(s, 32); err == nil {
			out := make([]float32, len(v))
			for i := range v {
				out[i] = float32(v[i])
			}
			result = out
		}
	case []float64:
		result, err = parseFloatList(s, 64)
	default:
		slog.Error("unsupported config type", "value", s)
		return zero, false
	}

	if err != nil {
		slog.Error("failed to parse config value", "value", s, "error", err)
		return zero, false
	}
	return result.(T), true
}

// splitList splits a comma-separated string and trims each element.
func splitList(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, len(parts))
	for i, p := range parts {
		out[i] = strings.TrimSpace(p)
	}
	return out
}

func parseIntList(s string, bitSize int) ([]int64, error) {
	parts := splitList(s)
	out := make([]int64, len(parts))
	for i, p := range parts {
		v, err := strconv.ParseInt(p, 10, bitSize)
		if err != nil {
			return nil, err
		}
		out[i] = v
	}
	return out, nil
}

func parseFloatList(s string, bitSize int) ([]float64, error) {
	parts := splitList(s)
	out := make([]float64, len(parts))
	for i, p := range parts {
		v, err := strconv.ParseFloat(p, bitSize)
		if err != nil {
			return nil, err
		}
		out[i] = v
	}
	return out, nil
}

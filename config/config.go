package config

import (
	"fmt"
	"log/slog"
	"path"
	"strings"

	"github.com/spf13/viper"
)

// ConfigMiddleware retrieve a secret value from a secret provider (see config/aws)
type ConfigMiddleware func(configKey string, configValue string) (string, error)

var _configMiddleware = func(_ string, configValue string) (string, error) {
	// default no-op
	return configValue, nil
}

// type ConfigMiddleware interface {
// 	ParseSecret(configKey string, configValue string) (string, error)
// }

func InitConfig(serviceName string) {
	InitConfigFull(serviceName, nil)
}

func InitConfigFull(serviceName string, configMiddleware ConfigMiddleware) {
	if configMiddleware != nil {
		_configMiddleware = configMiddleware
	}

	viper.SetConfigName(serviceName)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath(path.Join(".", "services", serviceName, "envs"))
	viper.AddConfigPath(path.Join(".", "envs"))
	viper.AddConfigPath(path.Join("..", "envs"))
	viper.AddConfigPath(path.Join("..", "..", "envs"))
	viper.AddConfigPath(path.Join("/", "etc", serviceName))
	err := viper.ReadInConfig()

	// empty service name force configuration using env variable only
	if serviceName != "" {
		if err != nil {
			slog.Info("no config file found, relying on ENV var")
		}
	}

	viper.SetEnvPrefix(strings.ToUpper(serviceName))
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	SetDefaultLogger()
}

// MustGetString returns the value of a key from the configuration file
func MustGetString(key string) string {
	if !viper.IsSet(key) {
		panic(fmt.Sprintf("key [%s] not set", key))
	}

	value, err := _configMiddleware(key, viper.GetString(key))
	if err != nil {
		panic(fmt.Errorf("failed to retrieve secret value for %s: %w", key, err))
	}

	return value
}

func GetString(key string) string {
	valueDecrypted, err := _configMiddleware(key, viper.GetString(key))
	if err != nil {
		slog.Error("failed to retrieve secret value for %s: %w", "key", key, "error", err)
		return ""
	}
	return valueDecrypted
}

// GetStringOr returns the value of a key from the configuration file, or defaultValue
func GetStringOr(key string, defaultValue string) string {
	if !viper.IsSet(key) {
		return defaultValue
	}

	value, err := _configMiddleware(key, viper.GetString(key))
	if err != nil {
		slog.Error("failed to retrieve secret value for %s: %w", "key", key, "error", err)
		return defaultValue
	}

	return value
}

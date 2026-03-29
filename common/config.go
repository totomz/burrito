package common

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

func InitConfig(serviceName string) {
	InitConfigWithReload(serviceName, false)
}

func InitConfigWithReload(serviceName string, enableReload bool) {
	env := GetEnvironment()

	viper.SetConfigName(fmt.Sprintf("config-%s", env))
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath(fmt.Sprintf("./services/%s/envs", serviceName))
	viper.AddConfigPath("./envs")
	viper.AddConfigPath("../envs")
	viper.AddConfigPath("../../envs")
	viper.AddConfigPath("/etc/heero")
	err := viper.ReadInConfig()

	// empty service name force configuration using env variable only
	if serviceName != "" {
		if err != nil {
			panic(err)
		}
	}

	viper.SetEnvPrefix(strings.ToUpper(serviceName))
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if enableReload {
		slog.Info("enabling configuration reload")
		viper.OnConfigChange(func(e fsnotify.Event) {
			slog.Info("configuration file changed", "file", e.Name)
		})
		viper.WatchConfig()
	}

	SetDefaultLogger()
}

// LoadAWSConfig checks if valid IAM credentials are available, and falls back to the "heero-${env}" profile if not.
func LoadAWSConfig(ctx context.Context /*, optFns ...func(*config.LoadOptions) error*/) aws.Config {

	// AWS EKS inject the region as env.
	// Otherwise, the region is taken from the config-<env> file
	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		awsRegion = MustGetString("aws.region")
	}
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(awsRegion))
	if err != nil {
		panic(fmt.Sprintf("error loading default aws config: %v", err))
	}
	slog.Info("AWS config", "region", cfg.Region)

	awsSharedConfigProfile := GetString("aws.sharedConfigProfile")
	if awsSharedConfigProfile != "" && !IsCICD() {
		cfg, err = config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(awsSharedConfigProfile), config.WithRegion(awsRegion))
		if err != nil {
			panic(fmt.Sprintf("error loading default aws local config: %v", err))
		}
	}

	return cfg
}

func GetAwsSecret(ctx context.Context, secretKey string, secretValueKey string) string {

	cfg := LoadAWSConfig(ctx)
	slog.Info("AWS config", "region", cfg.Region)

	// Create Secrets Manager client
	secretName := MustGetString(secretKey)
	svc := secretsmanager.NewFromConfig(cfg)

	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	result, err := svc.GetSecretValue(ctx, input)
	if err != nil {
		panic(err)
	}

	var secretData map[string]string
	err = json.Unmarshal([]byte(*result.SecretString), &secretData)
	if err != nil {
		panic(err)
	}

	secretValue, exists := secretData[secretValueKey]
	if !exists {
		panic("'apikey' not present in the secret")
	}

	return secretValue
}

// MustGetString returns the value of a key from the configuration file
func MustGetString(key string) string {
	if !viper.IsSet(key) {
		panic(fmt.Sprintf("key [%s] not set", key))
	}
	return viper.GetString(key)
}

func GetString(key string) string {
	return viper.GetString(key)
}

// GetStringOr returns the value of a key from the configuration file, or defaultValue
func GetStringOr(key string, defaultValue string) string {
	if !viper.IsSet(key) {
		return defaultValue
	}
	return viper.GetString(key)
}

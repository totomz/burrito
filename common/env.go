package common

import (
	"fmt"
	"os"
)

type Environment string

const (
	EnvironmentLocal Environment = "local"
	EnvironmentHack  Environment = "hack"
	EnvironmentProd  Environment = "prod"
)

// GetenvOrDefault retrieves the value of the environment variable named by the key.
// If the value is the empty string, returns defaultValue
func GetenvOrDefault(key, defaultValue string) string {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	return v
}

func SetEnvironment(env string) {
	err := os.Setenv("ENV", env)
	if err != nil {
		panic(err)
	}
}

func GetEnvironment() Environment {
	env := os.Getenv("ENV")
	if env == "" {
		env = "local"
	}

	switch env {
	case "local":
		return EnvironmentLocal
	case "hack":
		return EnvironmentHack
	case "prod":
		return EnvironmentProd
	default:
		panic(fmt.Sprintf("unsupported ENV: [%s]", env))
	}

}

func GetEnvHackProdOnly() Environment {
	env := GetEnvironment()
	if env == EnvironmentLocal {
		env = EnvironmentHack
	}
	return env
}

func IsCICD() bool {
	env := os.Getenv("CI")
	return env == "true"
}

func IsKube() bool {
	return len(os.Getenv("KUBERNETES_SERVICE_PORT")) > 0
}

// GetenvOrDie retrieves the value of the environment variable named by the key.
// This function panic() if the value match the empty string
func GetenvOrDie(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("missing required env var %s", key))
	}
	return v
}

//
// // GetEnv retrieves the value of the environment variable named by the key
// // if defined, otherwise return an error
// func GetEnv(k string) (string, error) {
// 	v := os.Getenv(k)
// 	if v == "" {
// 		return "", fmt.Errorf("%s environment variable not set", k)
// 	}
// 	return v, nil
// }
//
// func PanicIfErr(err error) {
// 	if err != nil {
// 		panic(err)
// 	}
// }

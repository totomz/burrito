package burrito_common

import (
	"fmt"
	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/burrito/")

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
}

func MustGetString(key string) string {
	dio := viper.GetString(key)
	if !viper.InConfig(key) {
		panic(fmt.Sprintf("missing configuration KEY %s in configuration storage %s", key, viper.ConfigFileUsed()))
	}
	return dio
}

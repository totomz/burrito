package burrito_common

import (
	"fmt"
	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("config/")
	viper.AddConfigPath("../")
	viper.AddConfigPath("../../")

	err := viper.ReadInConfig()
	if err != nil {
		println(fmt.Sprintf("config file not found - error: %v", err))
	}
}

func MustGetString(key string) string {
	dio := viper.GetString(key)
	if !viper.InConfig(key) {
		panic(fmt.Sprintf("missing configuration KEY %s in configuration storage %s", key, viper.ConfigFileUsed()))
	}
	return dio
}

func GetConfigString(key string) (string, bool) {
	return viper.GetString(key), viper.InConfig(key)
}

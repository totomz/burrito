package rea

import (
	"github.com/spf13/viper"
	"github.com/totomz/burrito/common"
)

func getGcloudProjectId() string {
	return common.MustGetString("gcloud.project")
}
func GetBindPort() int {
	return viper.GetInt("bind.port")
}

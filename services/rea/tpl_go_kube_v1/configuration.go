package rea

import (
	"github.com/totomz/burrito/v2/common"
)

func getGcloudProjectId() string {
	return common.MustGet[string]("gcloud.project")
}
func GetBindPort() int {
	return common.MustGet[int]("bind.port")
}

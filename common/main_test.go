package common

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	InitConfig("")
	os.Exit(m.Run())
}

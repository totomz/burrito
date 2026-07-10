package common

import (
	"log/slog"
	"testing"
)

func TestLogLine(t *testing.T) {
	SetDefaultLogger()
	slog.Info("ciao")
}

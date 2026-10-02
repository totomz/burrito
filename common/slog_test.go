package common

import (
	"log/slog"
	"testing"
)

func TestLogLine(t *testing.T) {
	SetDefaultLogger(nil)
	slog.Info("ciao")
}

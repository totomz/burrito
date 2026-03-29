// Package common slog implements the third iteration towards a logging framework that makes our CTO happy
// The current idea is that the Golang slog package is great, we only need some cosmetics to always log some
// values from the context
package common

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

var (
	// CtxRequestId is a key that must be used to attach a request id to a context.
	// The request id is always logged
	CtxRequestId = "__CTX_REQUEST_ID"

	// CtxDeviceId is a key that must be used to attach a device id to a context.
	// The device id is always logged
	CtxDeviceId = "__CTX_DEVICE_ID"

	// CtxRequestUser user id added to each log, if present in the context
	CtxRequestUser = "__CTX_REQUEST_USER"
)

func SetDefaultLogger() {
	slog.SetDefault(slog.New(NewCtxLogHandler()))
}

type CtxLogHandler struct {
	next slog.Handler
}

func NewCtxLogHandler() *CtxLogHandler {
	return &CtxLogHandler{
		next: slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true, ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Remove time.
			if a.Key == slog.TimeKey && len(groups) == 0 {
				return slog.Attr{}
			}

			if a.Key == slog.LevelKey && len(groups) == 0 {
				a.Value = slog.StringValue(fmt.Sprintf("%s", a.Value))
			}

			if a.Key == slog.SourceKey {
				source := a.Value.Any().(*slog.Source)
				source.File = shortenFilePath(source.File)
			}
			return a
		}}),
	}
}

func (h *CtxLogHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (h *CtxLogHandler) Handle(ctx context.Context, record slog.Record) error {
	if !IsKube() {
		return h.next.Handle(ctx, record)
	}

	requestId := ctx.Value(CtxRequestId)
	if requestId != nil {
		record.AddAttrs(slog.Any("request_id", requestId))
	}

	deviceId := ctx.Value(CtxDeviceId)
	if deviceId != nil {
		record.AddAttrs(slog.Any("device_id", deviceId))
	}

	userId := ctx.Value(CtxRequestUser)
	if userId != nil && userId != "" {
		userAlreadyInTheAttributes := false
		record.Attrs(func(attr slog.Attr) bool {
			if attr.Key == "user" {
				userAlreadyInTheAttributes = true
				return false
			}
			return true
		})
		if !userAlreadyInTheAttributes {
			user, isHeeroUser := userId.(User)
			if isHeeroUser {
				record.AddAttrs(slog.Any("user", user.UserID()))
			} else {
				record.AddAttrs(slog.Any("user_id", userId))
			}
		}
	}

	return h.next.Handle(ctx, record)
}

// WithAttrs returns a new handler with the provided attributes.
func (h *CtxLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &CtxLogHandler{
		next: h.next.WithAttrs(attrs),
	}
}

// WithGroup returns a new handler with the provided group name.
func (h *CtxLogHandler) WithGroup(name string) slog.Handler {
	return &CtxLogHandler{
		next: h.next.WithGroup(name),
	}
}

func shortenFilePath(path string) string {

	if len(path) < 9 {
		return path
	}
	if !strings.Contains(path, "/services") {
		return path
	}
	parts := strings.Split(path, "/services")
	return "services" + parts[1]
}

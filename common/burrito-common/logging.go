package burrito_common

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path"
	"time"
)

var (
	// CtxRequestId is a key that must be used to attach a request id to a context.
	// The request idd is always logged
	CtxRequestId = "__CTX_REQUEST_ID"
)

func MakeStdout(ctx context.Context) *slog.Logger {
	var stdout = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     nil,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				// logs in kube already have a timestamp attached to them
				// remoive it from the logline to keep it short
				isKube := os.Getenv("KUBERNETES_PORT")
				if isKube != "" {
					return slog.Attr{}
				}

				a.Key = "time"
				a.Value = slog.StringValue(time.Now().Format("2006-01-02 15:04:05Z07"))
			}

			if a.Key == slog.SourceKey {
				s := a.Value.Any().(*slog.Source)
				// the source file is logged between a blank and a column `\s<file>:`
				// so Intellij recognize it in the logs and we can click and go to the log line
				a.Value = slog.StringValue(fmt.Sprintf(" %s:%v:", path.Base(s.File), s.Line))
			}

			return a
		},
	}))

	reqUid := ctx.Value(CtxRequestId)
	if reqUid != "" {
		return stdout.With("req_id", reqUid)
	}

	return stdout
}

package slogdriver

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/karl-gustav/slogdriver/google"
)

var cloudProjectID string

const traceContextKey = "google-cloud-trace-id"

type cloudHandler struct{ slog.Handler }

func NewCloudHandler(projectID string, level slog.Level, optionalWriter ...io.Writer) *cloudHandler {
	cloudProjectID = projectID
	if len(optionalWriter) == 0 {
		optionalWriter[0] = os.Stderr
	}
	return &cloudHandler{Handler: slog.NewJSONHandler(optionalWriter[0], &slog.HandlerOptions{
		AddSource: true,
		Level:     level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.MessageKey {
				a.Key = "message"
			} else if a.Key == slog.LevelKey {
				a.Key = "severity"
			} else if a.Key == slog.SourceKey {
				a.Key = "logging.googleapis.com/sourceLocation"
			}
			return a
		},
	})}
}

func WithTraceContext(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var trace string
		traceHeader := r.Header.Get("X-Cloud-Trace-Context")
		traceParts := strings.Split(traceHeader, "/")
		if len(traceParts) > 0 && len(traceParts[0]) > 0 {
			trace = fmt.Sprintf("projects/%s/traces/%s", cloudProjectID, traceParts[0])
		}
		h.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), traceContextKey, trace)))
	})
}

func (h *cloudHandler) Handle(ctx context.Context, rec slog.Record) error {
	rec = rec.Clone()
	trace := traceFromContext(ctx)
	if trace != "" {
		// Add trace ID to the record so it is correlated with the Cloud Run request log
		// See https://cloud.google.com/trace/docs/trace-log-integration
		rec.Add("logging.googleapis.com/trace", slog.StringValue(trace))
	}
	if rec.Level == slog.LevelError {
		// See https://cloud.google.com/error-reporting/docs/formatting-error-messages#log-text
		rec.Add("@type", "type.googleapis.com/google.devtools.clouderrorreporting.v1beta1.ReportedErrorEvent")
	}
	// See https://cloud.google.com/error-reporting/docs/formatting-error-messages#reported-error-example
	rec.Add(slog.Group("serviceContext", slog.String("service", google.GetServiceName())))
	rec.Add("timestamp", time.Now())
	return h.Handler.Handle(ctx, rec)
}

func traceFromContext(ctx context.Context) string {
	trace := ctx.Value(traceContextKey)
	if trace == nil {
		return ""
	}
	return trace.(string)
}

func (h *cloudHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &cloudHandler{h.Handler.WithAttrs(attrs)}
}

func (h *cloudHandler) WithGroups(name string) slog.Handler {
	return &cloudHandler{h.Handler.WithGroup(name)}
}

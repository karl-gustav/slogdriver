package slogdriver

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

var cloudProjectID string

const traceContextKey = "google-cloud-trace-id"

type cloudHandler struct{ slog.Handler }

func NewCloudHandler(projectID string) *cloudHandler {
	cloudProjectID = projectID
	return &cloudHandler{Handler: slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.MessageKey {
				a.Key = "message"
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
	trace := traceFromContext(ctx)
	if trace != "" {
		rec = rec.Clone()
		// Add trace ID to the record so it is correlated with the Cloud Run request log
		// See https://cloud.google.com/trace/docs/trace-log-integration
		rec.Add("logging.googleapis.com/trace", slog.StringValue(trace))
	}
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

package slogdriver

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"time"
)

const (
	red     = "\033[0;31m"
	yellow  = "\033[0;33m"
	blue    = "\033[0;34m"
	magenta = "\033[0;35m"
	cyan    = "\033[0;36m"

	reset = "\033[0m"
)

type localHandler struct {
	slog.Handler
	internalWriter io.Writer
}

type internalWriter struct {
	externalWriter io.Writer
}

func NewLocalHandler(level slog.Level, optionalWriter ...io.Writer) *localHandler {
	var externalWriter io.Writer = os.Stderr

	if len(optionalWriter) != 0 {
		externalWriter = optionalWriter[0]
	}
	internalWriterImpl := internalWriter{externalWriter}

	return &localHandler{
		Handler: slog.NewJSONHandler(&internalWriterImpl, &slog.HandlerOptions{
			Level: level,
		}),
	}
}

func (i *internalWriter) Write(p []byte) (int, error) {
	var input map[string]any
	var buf bytes.Buffer
	err := json.Unmarshal(p, &input)
	if err != nil {
		return 0, err
	}
	if timeString, ok := input["time"].(string); ok {
		ts, err := time.Parse(time.RFC3339Nano, timeString)
		if err != nil {
			return 0, err
		}
		buf.Write([]byte(ts.Format("[15:04:05.000]") + " "))
		delete(input, "time")
	}
	if level, ok := input["level"].(string); ok {
		switch level {
		case "DEBUG":
			level = magenta + level
		case "INFO":
			level = blue + level
		case "WARN":
			level = yellow + level
		case "ERROR":
			level = red + level
		}
		buf.Write([]byte(level + ": " + reset))
		delete(input, "level")
	}
	if message, ok := input["msg"].(string); ok {
		buf.Write([]byte(cyan + message + " " + reset))
		delete(input, "msg")
	}
	j, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		return 0, err
	}
	buf.Write(j)
	_, err = i.externalWriter.Write(buf.Bytes())
	if err != nil {
		return 0, err
	}
	return buf.Len(), nil
}

func (h *localHandler) Handle(ctx context.Context, r slog.Record) error {
	return h.Handler.Handle(ctx, r)
}

func (h *localHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &localHandler{h.Handler.WithAttrs(attrs), h.internalWriter}
}

func (h *localHandler) WithGroup(name string) slog.Handler {
	return &localHandler{h.Handler.WithGroup(name), h.internalWriter}
}

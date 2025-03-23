package slogdriver

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"os"
)

const (
	red     = "\033[0;31m"
	yellow  = "\033[0;33m"
	blue    = "\033[0;34m"
	magenta = "\033[0;35m"
	cyan    = "\033[0;36m"

	reset = "\033[0m"
)

type localHandler struct{ slog.Handler }

func NewLocalHandler() *localHandler {
	return &localHandler{
		Handler: slog.NewJSONHandler(os.Stderr, nil),
	}
}

func (h *localHandler) Handle(ctx context.Context, r slog.Record) error {
	level := r.Level.String() + ":"

	switch r.Level {
	case slog.LevelDebug:
		level = magenta + level + reset
	case slog.LevelInfo:
		level = blue + level + reset
	case slog.LevelWarn:
		level = yellow + level + reset
	case slog.LevelError:
		level = red + level + reset
	}

	fields := make(map[string]any, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = a.Value.Any()

		return true
	})

	b, err := json.MarshalIndent(fields, "", "  ")
	if err != nil {
		return err
	}

	timeStr := r.Time.Format("[15:05:05.000]")
	msg := cyan + r.Message + reset

	log.New(os.Stderr, "", 0).Println(timeStr, level, msg, string(b))

	return nil
}

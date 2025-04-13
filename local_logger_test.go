package slogdriver_test

import (
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/karl-gustav/slogdriver"
)

type ListWriter struct {
	store []string
}

func (w *ListWriter) Write(p []byte) (n int, err error) {
	w.store = append(w.store, string(p))
	return len(p), nil
}

func TestLocalHandler(t *testing.T) {
	writer := ListWriter{}

	logger := slog.New(slogdriver.NewLocalHandler(slog.LevelDebug, &writer))

	logger.Debug("debug message", slog.String("debugKey", "debugValue"))
	logger.Info("info message", slog.String("infoKey", "infoValue"))
	logger.Warn("warn message", slog.String("warnKey", "warnValue"))
	logger.Error("error message", slog.String("errorKey", "errorValue"))
	logger.WithGroup("group1").Info("group message")
	logger.With("err", "error1").Error("error message", slog.String("errorKey", "errorValue"))
	time.Sleep(1 * time.Second)

	lines := writer.store
	for i, val := range lines {
		fmt.Printf("line %d: %v\n", i, val)
	}
	line := lines[0]
	if !strings.Contains(line, "debug message") {
		t.Errorf("output missing 'debug message':\n%s", line)
	}
	if !strings.HasSuffix(line, "\n") {
		t.Errorf("line missing newline suffix:\n%s", line)
	}
	line = lines[1]
	if !strings.Contains(line, "info message") {
		t.Errorf("output missing 'info message':\n%s", line)
	}
	line = lines[2]
	if !strings.Contains(line, "warn message") {
		t.Errorf("output missing 'warn message':\n%s", line)
	}
	line = lines[3]
	if !strings.Contains(line, "error message") {
		t.Errorf("output missing 'error message':\n%s", line)
	}
	line = lines[4]
	if !strings.Contains(line, "INFO:") {
		t.Errorf("output missing 'INFO:':\n%s", line)
	}
	line = lines[5]
	if !strings.Contains(line, "error1") {
		t.Errorf("output missing 'error1':\n%s", line)
	}
}

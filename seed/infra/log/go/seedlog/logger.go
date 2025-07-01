package seedlog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"time"
)

var colors = map[slog.Level]string{
	slog.LevelDebug: "\x1b[1;35m",
	slog.LevelInfo:  "\x1b[1;34m",
	slog.LevelWarn:  "\x1b[1;33m",
	slog.LevelError: "\x1b[1;31m",
}

var activeLevel slog.LevelVar

func SetLevel(level slog.Level) {
	activeLevel.Set(level)
}

type ColoredHandler struct {
	w io.Writer
}

func (h *ColoredHandler) Enabled(_ context.Context, l slog.Level) bool {
	return l >= activeLevel.Level()
}

func (h *ColoredHandler) Handle(ctx context.Context, r slog.Record) error {
	levelChar := rune('U')
	levelString := []rune(r.Level.String())
	if len(levelString) > 0 {
		levelChar = levelString[0]
	}
	file := ""
	line := 0
	fn := runtime.FuncForPC(r.PC)
	if fn != nil {
		file, line = fn.FileLine(r.PC)
	}
	message := fmt.Sprintf("%s%s %c %s[%d]\x1b[0m %s\n",
		colors[r.Level], r.Time.Format(time.RFC3339), levelChar, file, line-1, r.Message)
	_, err := h.w.Write([]byte(message))
	if err != nil {
		return err
	}
	return nil
}

func (h *ColoredHandler) WithAttrs(as []slog.Attr) slog.Handler {
	// Do not support attrs
	return h
}

func (h *ColoredHandler) WithGroup(name string) slog.Handler {
	// Do not support groups
	return h
}

func NewColoredHandler(w io.Writer) *ColoredHandler {
	return &ColoredHandler{
		w: w,
	}
}

func init() {
	activeLevel.Set(slog.LevelInfo)
	slog.SetDefault(slog.New(&ColoredHandler{w: os.Stderr}))
}

func Debugf(format string, args ...any) {
	logger := slog.Default()
	if !logger.Enabled(context.Background(), slog.LevelDebug) {
		return
	}
	pcs := [1]uintptr{}
	runtime.Callers(2, pcs[:]) // skip [Callers, Infof]
	r := slog.NewRecord(time.Now(), slog.LevelDebug, fmt.Sprintf(format, args...), pcs[0])
	_ = logger.Handler().Handle(context.Background(), r)
}

func Infof(format string, args ...any) {
	logger := slog.Default()
	if !logger.Enabled(context.Background(), slog.LevelInfo) {
		return
	}
	pcs := [1]uintptr{}
	runtime.Callers(2, pcs[:]) // skip [Callers, Infof]
	r := slog.NewRecord(time.Now(), slog.LevelInfo, fmt.Sprintf(format, args...), pcs[0])
	_ = logger.Handler().Handle(context.Background(), r)
}

func Warnf(format string, args ...any) {
	logger := slog.Default()
	if !logger.Enabled(context.Background(), slog.LevelWarn) {
		return
	}
	pcs := [1]uintptr{}
	runtime.Callers(2, pcs[:]) // skip [Callers, Infof]
	r := slog.NewRecord(time.Now(), slog.LevelWarn, fmt.Sprintf(format, args...), pcs[0])
	_ = logger.Handler().Handle(context.Background(), r)
}

func Errorf(format string, args ...any) {
	logger := slog.Default()
	if !logger.Enabled(context.Background(), slog.LevelError) {
		return
	}
	pcs := [1]uintptr{}
	runtime.Callers(2, pcs[:]) // skip [Callers, Infof]
	r := slog.NewRecord(time.Now(), slog.LevelError, fmt.Sprintf(format, args...), pcs[0])
	_ = logger.Handler().Handle(context.Background(), r)
}

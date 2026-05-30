package logger

import (
	"context"
	"io"
	"log/slog"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

type multiHandler struct {
	handlers []slog.Handler
}

func (h *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, r.Level) {
			if err := handler.Handle(ctx, r); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithAttrs(attrs)
	}
	return &multiHandler{handlers: handlers}
}

func (h *multiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithGroup(name)
	}
	return &multiHandler{handlers: handlers}
}

func New() *slog.Logger {
	_ = os.MkdirAll("logs", 0755)

	appWriter := &lumberjack.Logger{
		Filename:   "logs/billing-app.log",
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   true,
	}

	errorWriter := &lumberjack.Logger{
		Filename:   "logs/billing-error.log",
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   true,
	}

	appHandler := slog.NewJSONHandler(
		io.MultiWriter(os.Stdout, appWriter),
		&slog.HandlerOptions{Level: slog.LevelDebug},
	)

	errorHandler := slog.NewJSONHandler(
		errorWriter,
		&slog.HandlerOptions{Level: slog.LevelError},
	)

	return slog.New(&multiHandler{
		handlers: []slog.Handler{
			appHandler,
			errorHandler,
		},
	})
}
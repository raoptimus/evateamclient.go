package slogadapter

import (
	"context"
	"log/slog"
)

type Slog struct {
	logger *slog.Logger
}

func New(logger *slog.Logger) *Slog {
	if logger == nil {
		logger = slog.Default()
	}
	return &Slog{logger: logger}
}

func (a *Slog) Debug(ctx context.Context, msg string, args ...any) {
	a.logger.DebugContext(ctx, msg, args...)
}

func (a *Slog) Info(ctx context.Context, msg string, args ...any) {
	a.logger.InfoContext(ctx, msg, args...)
}

func (a *Slog) Warn(ctx context.Context, msg string, args ...any) {
	a.logger.WarnContext(ctx, msg, args...)
}

func (a *Slog) Error(ctx context.Context, msg string, args ...any) {
	a.logger.ErrorContext(ctx, msg, args...)
}

func (a *Slog) WithAttrs(attrs ...slog.Attr) *Slog {
	return &Slog{
		logger: a.logger.With(attrsToAny(attrs)...),
	}
}

func (a *Slog) WithGroup(name string) *Slog {
	return &Slog{
		logger: a.logger.WithGroup(name),
	}
}

func attrsToAny(attrs []slog.Attr) []any {
	result := make([]any, len(attrs))
	for i, attr := range attrs {
		result[i] = attr
	}
	return result
}

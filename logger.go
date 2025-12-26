package evateamclient

import (
	"context"
)

// Logger defines the logging interface.
// Compatible with slog, logrus, zap, and zerolog through adapters.
type Logger interface {
	Debug(ctx context.Context, msg string, args ...any)
	Info(ctx context.Context, msg string, args ...any)
	Warn(ctx context.Context, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
}

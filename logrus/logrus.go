package logrus

import (
	"context"

	"github.com/sirupsen/logrus"
)

type Logrus struct {
	logger *logrus.Logger
}

func New(logger *logrus.Logger) *Logrus {
	return &Logrus{logger: logger}
}

func (a *Logrus) Debug(ctx context.Context, msg string, args ...any) {
	fields := a.argsToFields(args)
	a.logger.WithContext(ctx).WithFields(fields).Debug(msg)
}

func (a *Logrus) Info(ctx context.Context, msg string, args ...any) {
	fields := a.argsToFields(args)
	a.logger.WithContext(ctx).WithFields(fields).Info(msg)
}

func (a *Logrus) Warn(ctx context.Context, msg string, args ...any) {
	fields := a.argsToFields(args)
	a.logger.WithContext(ctx).WithFields(fields).Warn(msg)
}

func (a *Logrus) Error(ctx context.Context, msg string, args ...any) {
	fields := a.argsToFields(args)
	a.logger.WithContext(ctx).WithFields(fields).Error(msg)
}

func (a *Logrus) argsToFields(args []any) logrus.Fields {
	fields := make(logrus.Fields, len(args))
	for i := 0; i < len(args)-1; i += 2 {
		if key, ok := args[i].(string); ok {
			fields[key] = args[i+1]
		}
	}
	return fields
}

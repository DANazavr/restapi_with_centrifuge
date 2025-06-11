package log

import (
	"context"

	"github.com/DANazavr/RATest/internal/common/meta"
	"github.com/sirupsen/logrus"
)

type Log struct {
	ctx       context.Context
	entry     *logrus.Entry
	component string
}

type LogConfig struct {
	Component string `default:"RATest"`
	LogLevel  string `default:"info"`
}

func NewLog(ctx context.Context, config *LogConfig) *Log {
	logger := logrus.New()
	if logLevel, err := logrus.ParseLevel(config.LogLevel); err == nil {
		logger.SetLevel(logLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel) // Default to Info level if parsing fails
		logger.Warnf("Invalid log level '%s', defaulting to 'info'", config.LogLevel)
	}

	return &Log{
		ctx:       ctx,
		entry:     logger.WithFields(logrus.Fields{"component": config.Component}),
		component: config.Component,
	}
}

func (l *Log) WithComponent(component string) *Log {
	newEntry := l.entry.WithField("component", component)
	return &Log{
		entry:     newEntry,
		component: component,
	}
}

func (l *Log) WithField(key string, value interface{}) *Log {
	return &Log{
		entry:     l.entry.WithField(key, value),
		component: l.component,
	}
}

func (l *Log) WithFields(fields logrus.Fields) *Log {
	return &Log{
		entry:     l.entry.WithFields(fields),
		component: l.component,
	}
}

func (l *Log) logWithContext(ctx context.Context) *logrus.Entry {
	if ctx == nil {
		return l.entry
	}

	entry := l.entry

	if val, ok := ctx.Value(meta.RequestIDKey).(int64); ok {
		entry = l.entry.WithField("requestID", val)
	}

	if val, ok := ctx.Value(meta.UserIDKey).(int64); ok {
		entry = l.entry.WithField("userID", val)
	}

	return entry
}

func (l *Log) Debug(ctx context.Context, args ...interface{}) {
	l.logWithContext(ctx).Debug(args...)
}

func (l *Log) Info(ctx context.Context, args ...interface{}) {
	l.logWithContext(ctx).Info(args...)
}

func (l *Log) Warn(ctx context.Context, args ...interface{}) {
	l.logWithContext(ctx).Warn(args...)
}

func (l *Log) Error(ctx context.Context, args ...interface{}) {
	l.logWithContext(ctx).Error(args...)
}

func (l *Log) Fatal(ctx context.Context, args ...interface{}) {
	l.logWithContext(ctx).Fatal(args...)
}

func (l *Log) Panic(ctx context.Context, args ...interface{}) {
	l.logWithContext(ctx).Panic(args...)
}

func (l *Log) Debugf(ctx context.Context, format string, args ...interface{}) {
	l.logWithContext(ctx).Debugf(format, args...)
}

func (l *Log) Infof(ctx context.Context, format string, args ...interface{}) {
	l.logWithContext(ctx).Infof(format, args...)
}

func (l *Log) Warnf(ctx context.Context, format string, args ...interface{}) {
	l.logWithContext(ctx).Warnf(format, args...)
}

func (l *Log) Errorf(ctx context.Context, format string, args ...interface{}) {
	l.logWithContext(ctx).Errorf(format, args...)
}

func (l *Log) Fatalf(ctx context.Context, format string, args ...interface{}) {
	l.logWithContext(ctx).Fatalf(format, args...)
}

func (l *Log) Panicf(ctx context.Context, format string, args ...interface{}) {
	l.logWithContext(ctx).Panicf(format, args...)
}

package logrus

import (
	"fmt"
	"io"
	"strings"

	"github.com/appnaconda/gohan/logger"
	"github.com/sirupsen/logrus"
)

type logrusLogger struct {
	logger *logrus.Entry
}

func (l *logrusLogger) Info(args ...interface{}) {
	l.logger.Info(args...)
}

func (l *logrusLogger) Infof(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}

func (l *logrusLogger) Debug(args ...interface{}) {
	l.logger.Debug(args...)
}

func (l *logrusLogger) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}

func (l *logrusLogger) Warn(args ...interface{}) {
	l.logger.Warn(args...)
}

func (l *logrusLogger) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}

func (l *logrusLogger) Error(args ...interface{}) {
	l.logger.Error(args...)
}

func (l *logrusLogger) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}

func (l *logrusLogger) Fatal(args ...interface{}) {
	l.logger.Fatal(args...)
}

func (l *logrusLogger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatalf(format, args...)
}

func (l *logrusLogger) SetLevel(level logger.Level) error {
	var err error
	l.logger.Logger.Level, err = logrus.ParseLevel(level.String())
	return err
}

func (l *logrusLogger) SetOutput(output io.Writer) {
	l.logger.Logger.Out = output
}

func (l *logrusLogger) SetOutputFormat(f logger.Format) error {
	switch strings.ToLower(f.String()) {
	case "text":
		l.logger.Logger.Formatter = &logrus.TextFormatter{}
	case "json":
		l.logger.Logger.Formatter = &logrus.JSONFormatter{}
	default:
		return fmt.Errorf("invalid log output format: %s. JSON will be used by default", f.String())
	}

	return nil
}

func (l *logrusLogger) With(fields logger.Fields) logger.Logger {
	f := logrus.Fields{}

	for k, v := range fields {
		f[k] = v
	}
	return &logrusLogger{
		l.logger.WithFields(f),
	}
}

func New(opts ...logger.Option) logger.Logger {
	logger := &logrusLogger{
		logger: logrus.NewEntry(logrus.New()),
	}

	for _, opt := range opts {
		opt.Apply(logger)
	}

	return logger
}

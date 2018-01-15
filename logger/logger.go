// This package contains the Logger interface definition.
package logger

import (
	"io"
)

type Option interface {
	Apply(logger Logger) error
}

type Fields map[string]interface{}

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
)

func (s Level) String() string {
	return levelName[s]
}

var levelName = [...]string{
	"DEBUG",
	"INFO",
	"WARN",
	"ERROR",
	"FATAL",
}

var levelValue = map[string]Level{
	"DEBUG": DEBUG,
	"INFO":  INFO,
	"WARN":  WARN,
	"ERROR": ERROR,
	"FATAL": FATAL,
}

func ParseLevel(l string) (Level, bool) {
	v, ok := levelValue[l]
	return v, ok
}

type Format int

func (f Format) String() string {
	return formatName[f]
}

const (
	JSON_FORMAT Format = iota
	TEXT_FORMAT
)

var formatName = [...]string{
	"JSON",
	"TEXT",
}

var formatValue = map[string]Format{
	"JSON": JSON_FORMAT,
	"TEXT": TEXT_FORMAT,
}

func ParseFormat(f string) (Format, bool) {
	v, ok := formatValue[f]
	return v, ok
}

type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})

	Info(args ...interface{})
	Infof(format string, args ...interface{})

	Warn(args ...interface{})
	Warnf(format string, args ...interface{})

	Error(args ...interface{})
	Errorf(format string, args ...interface{})

	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})

	With(fields Fields) Logger

	SetLevel(level Level) error
	SetOutput(output io.Writer)
	SetOutputFormat(format Format) error
}

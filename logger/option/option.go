package option

import "github.com/appnaconda/gohan/logger"

func WithLevel(level logger.Level) logger.Option {
	return withLogLevel{level: level}
}

type withLogLevel struct {
	level logger.Level
}

func (lv withLogLevel) Apply(logger logger.Logger) error {
	logger.SetLevel(lv.level)
	return nil
}

func WithFormat(format logger.Format) logger.Option {
	return withLogFormat{format: format}
}

type withLogFormat struct {
	format logger.Format
}

func (f withLogFormat) Apply(logger logger.Logger) error {
	logger.SetOutputFormat(f.format)
	return nil
}

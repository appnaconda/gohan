package option

import (
	"github.com/appnaconda/gohan"
	"github.com/appnaconda/gohan/tracer/google"
)

func WithGoogleTracer() gohan.Option {
	return withGoogleTracer{}
}

type withGoogleTracer struct {
	name string
}

func (sn withGoogleTracer) Apply(s *gohan.Service) error {
	tracer, err := google.NewTracer(s.Context)
	if err != nil {
		return err
	}

	s.Tracer = tracer

	return nil
}

package option

import (
	"context"

	"github.com/appnaconda/gohan"
)

func WithName(name string) gohan.Option {
	return withServiceName{name: name}
}

type withServiceName struct {
	name string
}

func (sn withServiceName) Apply(s *gohan.Service) error {
	s.Name = sn.name
	return nil
}

func WithVersion(version string) gohan.Option {
	return withServiceVersion{version: version}
}

type withServiceVersion struct {
	version string
}

func (sv withServiceVersion) Apply(s *gohan.Service) error {
	s.Version = sv.version
	return nil
}

func WithContext(ctx context.Context) gohan.Option {
	return withContext{ctx: ctx}
}

type withContext struct {
	ctx context.Context
}

func (c withContext) Apply(s *gohan.Service) error {
	s.Context = c.ctx
	return nil
}

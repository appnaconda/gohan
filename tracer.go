package gohan

import (
	"context"
	"net/http"
)

type Tracer interface {
	GetSpan(ctx context.Context) Span
	HTTPHandler(http.Handler) http.Handler
}

type Span interface {
	NewChild(string) Span
	SetLabel(k, v string)
	Finish()
}

type nullTracer struct{}

func (nullTracer) GetSpan(ctx context.Context) Span {
	return nullSpan{}
}

type nullSpan struct{}

func (nullSpan) NewChild(string) Span {
	return nullSpan{}
}

func (nullSpan) SetLabel(k, v string) {}

func (nullSpan) Finish() {}

package google

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"cloud.google.com/go/trace"
	"github.com/appnaconda/gohan"
	"google.golang.org/api/option"
)

type tracer struct {
	client *trace.Client
}

type span struct {
	parent *trace.Span
}

func NewTracer(ctx context.Context) (gohan.Tracer, error) {
	projectID := os.Getenv("GCE_PROJECT_ID")
	if projectID == "" {
		return nil, fmt.Errorf("GCE_PROJECT_ID env variable is needed for google tracer")
	}

	var traceClient *trace.Client
	var err error

	credentialFile := os.Getenv("GCE_TRACE_CREDENTIAL_FILE")
	if credentialFile != "" {
		traceClient, err = trace.NewClient(ctx, projectID, option.WithCredentialsFile(credentialFile))
	} else {
		// trying with the default credential: GOOGLE_APPLICATION_CREDENTIALS env var
		traceClient, err = trace.NewClient(ctx, projectID)
	}

	if err != nil {
		return nil, err
	}

	return &tracer{
		client: traceClient,
	}, nil
}

func (t *tracer) HTTPHandler(h http.Handler) http.Handler {
	if t.client == nil {
		return h
	}

	return t.client.HTTPHandler(h)
}

func (tracer) GetSpan(ctx context.Context) gohan.Span {
	if ctx == nil {
		return span{parent: nil}
	}

	return span{parent: trace.FromContext(ctx)}
}

func (s span) NewChild(name string) gohan.Span {
	return span{parent: s.parent.NewChild(name)}
}

func (s span) SetLabel(k, v string) {
	s.parent.SetLabel(k, v)
}

func (s span) Finish() {
	s.parent.Finish()
}

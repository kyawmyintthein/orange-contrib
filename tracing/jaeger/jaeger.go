package jaeger

import (
	"context"
	"net/http"

	"github.com/opentracing/opentracing-go"
)

type JaegerTracer interface {
	IsEnabled() bool
	HttpClientTracer(context.Context, *http.Request, string) opentracing.Span
}

package jaegerx

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/kyawmyintthein/orange-contrib/logx"
	"github.com/kyawmyintthein/orange-contrib/optionx"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/uber/jaeger-client-go"
	jconfig "github.com/uber/jaeger-client-go/config"
)

const (
	HTTP_URL    = "http.url"
	REQUEST_URI = "request.uri"
)

type JaegerTracer interface {
	MiddlewareTracer(http.Handler) http.Handler
	HttpClientTracer(context.Context, *http.Request, string) opentracing.Span
	Close() error
	IsEnabled() bool
}

type jaegerClient struct {
	cfg    *JaegerCfg
	tracer opentracing.Tracer
	closer io.Closer
	logger logx.Logger
}

func New(cfg *JaegerCfg, opts ...optionx.Option) (Jaeger, error) {
	options := optionx.NewOptions(opts...)

	jaegerClient := jaegerClient{
		cfg: cfg,
	}

	jaegerConfiguration := &jconfig.Configuration{
		Sampler: &jconfig.SamplerConfig{
			Type:              cfg.SamplerType,
			Param:             cfg.SamplerParam,
			SamplingServerURL: cfg.SamplingServerURL,
		},
		Reporter: &jconfig.ReporterConfig{
			LogSpans:           cfg.LogSpans,
			CollectorEndpoint:  cfg.ReporterCollectorEndpoint,
			LocalAgentHostPort: cfg.LocalAgentPort,
		},
	}
	jaegerConfiguration.ServiceName = cfg.LocalServiceName
	jaegerConfiguration.Sampler.Type = cfg.SamplerType
	jaegerConfiguration.Sampler.Param = cfg.SamplerParam

	// set custom logger
	logger, ok := options.Context.Value(loggerKey{}).(logx.Logger)
	if logger != nil || ok {
		jaegerClient.logger = logger
	} else {
		jaegerClient.logger = logx.New(&logx.LogCfg{})
	}

	// set jaeger logger
	jaegerLogger, ok := options.Context.Value(jaegerLoggerKey{}).(jaeger.Logger)
	if jaegerLogger != nil || !ok {
		jaegerLogger = jaeger.StdLogger
		jaegerClient.logger.Info(context.Background(), "[CLJaeger] Jaeger is using jaeger.StdLogger")
	}

	tracer, closer, err := jaegerConfiguration.NewTracer(
		jconfig.Logger(jaegerLogger),
	)

	if err != nil {
		jaegerClient.cfg.Enabled = false
		jaegerClient.logger.Error(context.Background(), err, "[CLJaeger] failed to initiate Jaeger")
		return &jaegerClient, err
	}

	jaegerClient.tracer = tracer
	jaegerClient.closer = closer

	jaegerClient.logger.Info(context.Background(), "[CLJaeger] Jaeger is successfully initiated")
	return &jaegerClient, nil
}

func (jaegerClient *jaegerClient) HttpClientTracer(ctx context.Context, req *http.Request, operationName string) opentracing.Span {
	if !jaegerClient.cfg.Enabled {
		return nil
	}
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, jaegerClient.tracer, operationName)
	ext.SpanKindRPCClient.Set(span)
	ext.HTTPUrl.Set(span, req.URL.String())
	ext.HTTPMethod.Set(span, req.Method)
	for k, v := range req.Header {
		span.SetTag(fmt.Sprintf("request.header.%s", strings.ToLower(k)), v)
	}

	req = req.WithContext(ctx)
	span.Tracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header),
	)

	jaegerClient.logger.Info(ctx, "Span Injected")
	return span
}

func (jaegerClient *jaegerClient) MiddlewareTracer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !jaegerClient.cfg.Enabled {
			return
		}
		requestSpan := jaegerClient.newRequestSpan(r)
		r = r.WithContext(opentracing.ContextWithSpan(r.Context(), requestSpan))
		w = &statusCodeTracker{w, 200}

		//traceID := r.Header.Get("Circle-Trace-Id")
		//if traceID != "" {
		//	traceID = strings.Split(traceID, ":")[0]
		//	w.Header().Set("X-Trace-Id", traceID)
		//	// adds traceID to a context and get from it latter
		//	r = r.WithContext(jaegerService.tracer.WithContextValue(r.Context(), traceID))
		//}
		next.ServeHTTP(w, r)
		for k, v := range w.Header() {
			requestSpan.SetTag(fmt.Sprintf("response.header.%s", strings.ToLower(k)), v)
		}
		ext.HTTPStatusCode.Set(requestSpan, uint16(w.(*statusCodeTracker).status))
		defer requestSpan.Finish()
	})
}

func (jaegerClient *jaegerClient) newRequestSpan(r *http.Request) opentracing.Span {
	var span opentracing.Span
	operation := fmt.Sprintf("HTTP %s %s", r.Method, r.URL.String())
	carrier := opentracing.HTTPHeadersCarrier(r.Header)
	wireContext, err := jaegerClient.tracer.Extract(opentracing.HTTPHeaders, carrier)
	if err != nil {
		span = jaegerClient.tracer.StartSpan(operation)
	} else {
		span = jaegerClient.tracer.StartSpan(operation, opentracing.ChildOf(wireContext))
	}

	// it adds the trace ID to the http headers
	if err := span.Tracer().Inject(span.Context(), opentracing.HTTPHeaders, carrier); err != nil {
		ext.Error.Set(span, true)
	} else {
		r = r.WithContext(opentracing.ContextWithSpan(r.Context(), span))
	}

	ext.HTTPMethod.Set(span, r.Method)
	ext.HTTPUrl.Set(span, r.URL.String())
	for k, v := range r.Header {
		span.SetTag(fmt.Sprintf("request.header.%s", strings.ToLower(k)), v)
	}

	return span
}

func (jaegerClient *jaegerClient) Close() error {
	if !jaegerClient.cfg.Enabled {
		return nil
	}
	jaegerClient.cfg.Enabled = true
	return jaegerClient.closer.Close()
}

func (jaegerClient *jaegerClient) IsEnabled() bool {
	return jaegerClient.cfg.Enabled
}

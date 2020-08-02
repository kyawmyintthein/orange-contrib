package clientx

import (
	"context"
	"time"

	"github.com/kyawmyintthein/orange-contrib/logx"
	"github.com/kyawmyintthein/orange-contrib/optionx"
	"github.com/kyawmyintthein/orange-contrib/tracingx/jaegerx"
	"github.com/kyawmyintthein/orange-contrib/tracingx/newrelicx"
)

/*
	WithNewrelic - is to provide newrelic object to http client for API metric
*/
type newrelicTracerKey struct{}

func WithNewrelic(obj newrelicx.NewrelicTracer) optionx.Option {
	return func(o *optionx.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, newrelicTracerKey{}, obj)
	}
}

/*
	WithLogger - is to provide logger object http from your application to collect log information in single place.
*/
type loggerKey struct{}

func WithLogger(log logx.Logger) optionx.Option {
	return func(o *optionx.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, loggerKey{}, log)
	}
}

/*
	WithJaeger - is to provide jaeger tracer object http which can be used in http client for distributed tracing
*/
type jaegerTracerKey struct{}

func WithJaeger(a jaegerx.JaegerTracer) optionx.Option {
	return func(o *optionx.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, jaegerTracerKey{}, a)
	}
}

/*
	WithHeader - is to provide custom http header key and value to http client.
*/
type httpHeaderKey struct{}

func WithHeader(obj Header) optionx.Option {
	return func(o *optionx.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, httpHeaderKey{}, obj)
	}
}

/*
	WithOpName - is to provide operation name for each API call. This name will be used in metric and log to provide meaningful information.
				 Operation name should not contain space and case insensitive.
	For example;
		`GetUserProfileAPI` for GET::/users/{:id}
*/
type operationNameKey struct{}

func WithOpName(name string) optionx.Option {
	return func(o *optionx.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, operationNameKey{}, name)
	}
}

/*
	WithRetrySetting - is to provide retry configuration setting for each API call.
					   This will override the default retry setting from http clent's configuration
*/
type retrySettingKey struct{}

func WithRetrySetting(obj *RetryCfg) optionx.Option {
	return func(o *optionx.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, retrySettingKey{}, obj)
	}
}

type httpRequestTimeoutKey struct{}

func WithRequestTimeout(timeout time.Duration) option.Option {
	return func(o *option.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, httpRequestTimeoutKey{}, timeout)
	}
}

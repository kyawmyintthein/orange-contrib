package client

import (
	"context"
	"time"

	"github.com/kyawmyintthein/orange-contrib/cb"
	"github.com/kyawmyintthein/orange-contrib/logging"
	"github.com/kyawmyintthein/orange-contrib/option"
	"github.com/kyawmyintthein/orange-contrib/tracing/jaeger"
	"github.com/kyawmyintthein/orange-contrib/tracing/newrelic"
)

/*
	WithNewrelic - is to provide newrelic object to http client for API metric
*/
type newrelicTracerKey struct{}

func WithNewrelic(obj newrelic.NewrelicTracer) option.Option {
	return func(o *option.Options) {
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

func WithLogger(a logging.Logger) option.Option {
	return func(o *option.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, loggerKey{}, a)
	}
}

/*
	WithJaeger - is to provide jaeger tracer object http which can be used in http client for distributed tracing
*/
type jaegerTracerKey struct{}

func WithJaeger(a jaeger.JaegerTracer) option.Option {
	return func(o *option.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, jaegerTracerKey{}, a)
	}
}

/*
	WithCircuitBreaker - is to provide circuit breaker feature for http client
*/
type circuitBreakerKey struct{}

func WithCircuitBreaker(obj cb.CircuitBreaker) option.Option {
	return func(o *option.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, circuitBreakerKey{}, obj)
	}
}

/*
	WithHeader - is to provide custom http header key and value to http client.
*/
type httpHeaderKey struct{}

func WithHeader(obj Header) option.Option {
	return func(o *option.Options) {
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

func WithOpName(name string) option.Option {
	return func(o *option.Options) {
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

func WithRetrySetting(obj *RetryCfg) option.Option {
	return func(o *option.Options) {
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

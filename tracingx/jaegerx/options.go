package jaegerx

import (
	"context"

	"github.com/kyawmyintthein/orange-contrib/logx"
	"github.com/kyawmyintthein/orange-contrib/optionx"
	"github.com/uber/jaeger-client-go"
)

type jaegerLoggerKey struct{}

func WithJaegerLogger(a jaeger.Logger) optionx.Option {
	return func(o *optionx.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, jaegerLoggerKey{}, a)
	}
}

type loggerKey struct{}

func WithLogger(a logx.Logger) optionx.Option {
	return func(o *optionx.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, loggerKey{}, a)
	}
}

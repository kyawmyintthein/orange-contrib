package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/kyawmyintthein/orange-contrib/logging"
	"github.com/kyawmyintthein/orange-contrib/option"
	"github.com/kyawmyintthein/orange-contrib/tracing/jaeger"
	"github.com/kyawmyintthein/orange-contrib/tracing/newrelic"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
)

const (
	httpDeleteMethod string = "DELETE"
	httpGetMethod    string = "GET"
	httpPostMethod   string = "POST"
	httpPutMethod    string = "PUT"
	httpPatchMethod  string = "PATCH"
)

const (
	applicationJSON string = "application/json"
)

const (
	defaultRequestTimeout  time.Duration = 10
	defaultRetryAttempts   uint          = 3
	defaultBackOffDuration time.Duration = 100 * time.Millisecond
)

var (
	defaultExponentialBackOffDurations = []time.Duration{
		defaultBackOffDuration,
		time.Duration(200 * time.Millisecond),
		time.Duration(1 * time.Second),
	}
)

type Header map[string]string

type HttpClient interface {
	POST(context.Context, string, interface{}, ...option.Option) (*http.Response, error)
	PUT(context.Context, string, interface{}, ...option.Option) (*http.Response, error)
	PATCH(context.Context, string, interface{}, ...option.Option) (*http.Response, error)
	DELETE(context.Context, string, interface{}, ...option.Option) (*http.Response, error)
	GET(context.Context, string, ...option.Option) (*http.Response, error)
}

type httpClient struct {
	cfg *HttpClientCfg

	logger         logging.Logger
	jaegerTracer   jaeger.JaegerTracer
	newrelicTracer newrelic.NewrelicTracer
}

func DefaultHttpClient(cfg *HttpClientCfg, opts ...option.Option) HttpClient {
	options := option.NewOptions(opts...)

	httpClient := &httpClient{
		cfg: cfg,
	}

	//set newrelic
	newrelicTracer, ok := options.Context.Value(newrelicTracerKey{}).(newrelic.NewrelicTracer)
	if newrelicTracer != nil && ok {
		httpClient.newrelicTracer = newrelicTracer
	}

	// set jaeger
	jaeger, ok := options.Context.Value(jaegerTracerKey{}).(jaeger.JaegerTracer)
	if jaeger != nil || !ok {
		httpClient.jaegerTracer = jaeger
	}

	// set logger
	logger, ok := options.Context.Value(loggerKey{}).(logging.Logger)
	if logger != nil && ok {
		httpClient.logger = logger
	} else {
		httpClient.logger = logging.DefaultLogger()
	}

	return httpClient
}

func (httpClient *httpClient) DELETE(ctx context.Context, url string, payload interface{}, opts ...option.Option) (*http.Response, error) {
	options := option.NewOptions(opts...)
	retryConfig := httpClient.getRetrySetting(ctx, httpDeleteMethod, url)

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(httpDeleteMethod, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	operationName := httpClient.getOpNameFromOption(url, httpDeleteMethod, options)
	req.Header.Set("Content-Type", applicationJSON)
	httpClient.setHeaderFromOption(req, options)

	//TODO: improvement
	var span opentracing.Span
	if httpClient.jaegerTracer != nil && httpClient.jaegerTracer.IsEnabled() {
		span = httpClient.jaegerTracer.HttpClientTracer(ctx, req, operationName)
		defer span.Finish()
	}

	resp, err := httpClient.firstAttemptAndRetry(ctx, &retryConfig, req, operationName, options)
	if err != nil {
		return nil, err
	}

	//TODO: improvement
	if httpClient.jaegerTracer != nil && httpClient.jaegerTracer.IsEnabled() && span != nil {
		span.SetTag("http.response.status", resp.StatusCode)
		for k, v := range resp.Header {
			span.SetTag(fmt.Sprintf("http.response.header.%s", k), v)
		}
	}

	httpClient.logger.InfoKV(ctx, logging.KV{"URL": url, "Status": resp.Status, "Headers": resp.Header}, "[DELETE] Received response")
	return resp, err
}

func (httpClient *httpClient) setHeaderFromOption(req *http.Request, options option.Options) {
	header, ok := options.Context.Value(httpHeaderKey{}).(Header)
	if header == nil || !ok {
		return
	}
	for hk, hv := range header {
		req.Header.Set(hk, hv)
	}
}

func (httpClient *httpClient) getOpNameFromOption(url string, httpMethod string, options option.Options) string {
	opName, ok := options.Context.Value(operationNameKey{}).(string)
	if opName == "" || !ok {
		return fmt.Sprintf("%s::%s", httpMethod, url)
	}
	return opName
}

func (httpClient *httpClient) getRetrySetting(ctx context.Context, httpMethod string, url string) RetryCfg {
	retryConfig, ok := httpClient.getAPISpecificRetrySetting(ctx, httpMethod, url)
	if !ok {
		retryConfig = httpClient.cfg.DefaultRetrySetting
	}

	if uint(len(retryConfig.BackOffDurations)) < retryConfig.MaxRetryAttempts {
		backOffLen := len(retryConfig.BackOffDurations)
		if backOffLen == 0 {
			var i uint = 0
			for i < defaultRetryAttempts {
				retryConfig.BackOffDurations = append(retryConfig.BackOffDurations, defaultBackOffDuration)
				i++
			}
		} else {
			missingAttempts := retryConfig.MaxRetryAttempts - uint(backOffLen)
			lastBackoffDuration := retryConfig.BackOffDurations[backOffLen-1]
			var i uint = 0
			for i < missingAttempts {
				retryConfig.BackOffDurations = append(retryConfig.BackOffDurations, lastBackoffDuration)
				i++
			}
		}
	}
	return retryConfig
}

func (httpClient *httpClient) getAPISpecificRetrySetting(ctx context.Context, httpMethod string, url string) (RetryCfg, bool) {
	key := fmt.Sprintf("[%s]::/%s", httpMethod, url)
	retryConfig, ok := httpClient.cfg.APISpecificRetrySetting[key]
	return retryConfig, ok
}

func (httpClient *httpClient) firstAttemptAndRetry(ctx context.Context, retryConfig *RetryCfg, req *http.Request, operationName string, options option.Options) (*http.Response, error) {
	var count uint
	resp, err := httpClient.sendHttpRequest(ctx, req, operationName, options)
	if err != nil {
		if retryConfig.Enabled {
			return resp, err
		}

		for count < retryConfig.MaxRetryAttempts {
			resp, err := httpClient.sendHttpRequest(ctx, req, operationName, options)
			if err != nil {
				if count == retryConfig.MaxRetryAttempts-1 {
					return resp, err
				}

				backOffDuration := defaultBackOffDuration
				if uint(len(retryConfig.BackOffDurations)) >= count {
					backOffDuration = retryConfig.BackOffDurations[count]
				}

				time.Sleep(backOffDuration)
			} else {
				return resp, err
			}
			count++
		}
	}
	return resp, err
}

func (httpClient *httpClient) sendHttpRequest(ctx context.Context, req *http.Request, name string, options option.Options) (*http.Response, error) {
	client := http.Client{Transport: &nethttp.Transport{}}
	requestTimeout, ok := options.Context.Value(httpRequestTimeoutKey{}).(time.Duration)
	if !ok {
		requestTimeout = defaultRequestTimeout
	}
	client.Timeout = requestTimeout * time.Second

	if httpClient.newrelicTracer == nil || !httpClient.newrelicTracer.IsEnabled() {
		return client.Do(req)
	}

	es, err := httpClient.newrelicTracer.RecordExternalMetric(req, name)
	if err == nil {
		defer es.End()
	}

	response, err := client.Do(req)
	return response, err
}

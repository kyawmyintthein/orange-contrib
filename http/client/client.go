package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/kyawmyintthein/orange-contrib/cb"
	"github.com/kyawmyintthein/orange-contrib/errorx"
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

	err5xx = errorx.New(001, "Server_Return_5XX_StatusCode", "server return 5xx status code")
)

type Header map[string]string

type HttpClient interface {
	POST(context.Context, string, io.Reader, ...option.Option) (*http.Response, error)
	PUT(context.Context, string, io.Reader, ...option.Option) (*http.Response, error)
	PATCH(context.Context, string, io.Reader, ...option.Option) (*http.Response, error)
	DELETE(context.Context, string, io.Reader, ...option.Option) (*http.Response, error)
	GET(context.Context, string, ...option.Option) (*http.Response, error)
}

type httpClient struct {
	cfg *HttpClientCfg

	logger         logging.Logger
	jaegerTracer   jaeger.JaegerTracer
	newrelicTracer newrelic.NewrelicTracer
	circuitBreaker cb.CircuitBreaker
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

	// set circuit breaker object
	circuitBreaker, ok := options.Context.Value(circuitBreakerKey{}).(cb.CircuitBreaker)
	if circuitBreaker != nil || !ok {
		httpClient.circuitBreaker = circuitBreaker
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

func (httpClient *httpClient) GET(ctx context.Context, url string, opts ...option.Option) (*http.Response, error) {
	options := option.NewOptions(opts...)
	retryConfig := httpClient.getRetrySetting(ctx, httpGetMethod, url)
	operationName := httpClient.getOpNameFromOption(url, httpGetMethod, options)

	var resp *http.Response
	req, err := http.NewRequest(httpDeleteMethod, url, nil)
	if err != nil {
		return resp, err
	}

	req.Header.Set("Content-Type", applicationJSON)
	httpClient.setHeaderFromOption(req, options)

	//TODO: improvement
	var span opentracing.Span
	if httpClient.cfg.TurnOffJaeger || httpClient.jaegerTracer != nil && httpClient.jaegerTracer.IsEnabled() {
		span = httpClient.jaegerTracer.HttpClientTracer(ctx, req, operationName)
		defer span.Finish()
	}

	resp, err = httpClient.firstAttemptAndRetry(ctx, &retryConfig, req, operationName, options)
	if err != nil {
		return resp, err
	}

	//TODO: improvement
	if httpClient.jaegerTracer != nil && httpClient.jaegerTracer.IsEnabled() && span != nil {
		span.SetTag("http.response.status", resp.StatusCode)
		for k, v := range resp.Header {
			span.SetTag(fmt.Sprintf("http.response.header.%s", k), v)
		}
	}

	httpClient.logger.InfoKVf(ctx, logging.KV{"URL": url, "Status": resp.Status, "Headers": resp.Header}, "[%s] Received response", httpGetMethod)
	return resp, nil
}

func (httpClient *httpClient) POST(ctx context.Context, url string, body io.Reader, opts ...option.Option) (*http.Response, error) {
	options := option.NewOptions(opts...)
	retryConfig := httpClient.getRetrySetting(ctx, httpPostMethod, url)
	operationName := httpClient.getOpNameFromOption(url, httpPostMethod, options)

	var resp *http.Response
	req, err := http.NewRequest(httpDeleteMethod, url, body)
	if err != nil {
		return resp, err
	}

	req.Header.Set("Content-Type", applicationJSON)
	httpClient.setHeaderFromOption(req, options)

	//TODO: improvement
	var span opentracing.Span
	if httpClient.cfg.TurnOffJaeger || httpClient.jaegerTracer != nil && httpClient.jaegerTracer.IsEnabled() {
		span = httpClient.jaegerTracer.HttpClientTracer(ctx, req, operationName)
		defer span.Finish()
	}

	resp, err = httpClient.firstAttemptAndRetry(ctx, &retryConfig, req, operationName, options)
	if err != nil {
		return resp, err
	}

	//TODO: improvement
	if httpClient.jaegerTracer != nil && httpClient.jaegerTracer.IsEnabled() && span != nil {
		span.SetTag("http.response.status", resp.StatusCode)
		for k, v := range resp.Header {
			span.SetTag(fmt.Sprintf("http.response.header.%s", k), v)
		}
	}

	httpClient.logger.InfoKVf(ctx, logging.KV{"URL": url, "Status": resp.Status, "Headers": resp.Header}, "[%s] Received response", httpPostMethod)
	return resp, nil
}

func (httpClient *httpClient) PUT(ctx context.Context, url string, body io.Reader, opts ...option.Option) (*http.Response, error) {
	options := option.NewOptions(opts...)
	retryConfig := httpClient.getRetrySetting(ctx, httpPutMethod, url)
	operationName := httpClient.getOpNameFromOption(url, httpPutMethod, options)

	var resp *http.Response
	req, err := http.NewRequest(httpDeleteMethod, url, body)
	if err != nil {
		return resp, err
	}

	req.Header.Set("Content-Type", applicationJSON)
	httpClient.setHeaderFromOption(req, options)

	//TODO: improvement
	var span opentracing.Span
	if httpClient.cfg.TurnOffJaeger || httpClient.jaegerTracer != nil && httpClient.jaegerTracer.IsEnabled() {
		span = httpClient.jaegerTracer.HttpClientTracer(ctx, req, operationName)
		defer span.Finish()
	}

	resp, err = httpClient.firstAttemptAndRetry(ctx, &retryConfig, req, operationName, options)
	if err != nil {
		return resp, err
	}

	//TODO: improvement
	if httpClient.jaegerTracer != nil && httpClient.jaegerTracer.IsEnabled() && span != nil {
		span.SetTag("http.response.status", resp.StatusCode)
		for k, v := range resp.Header {
			span.SetTag(fmt.Sprintf("http.response.header.%s", k), v)
		}
	}

	httpClient.logger.InfoKVf(ctx, logging.KV{"URL": url, "Status": resp.Status, "Headers": resp.Header}, "[%s] Received response", httpPutMethod)
	return resp, nil
}

func (httpClient *httpClient) PATCH(ctx context.Context, url string, body io.Reader, opts ...option.Option) (*http.Response, error) {
	options := option.NewOptions(opts...)
	retryConfig := httpClient.getRetrySetting(ctx, httpPatchMethod, url)
	operationName := httpClient.getOpNameFromOption(url, httpPatchMethod, options)

	var resp *http.Response
	req, err := http.NewRequest(httpDeleteMethod, url, body)
	if err != nil {
		return resp, err
	}

	req.Header.Set("Content-Type", applicationJSON)
	httpClient.setHeaderFromOption(req, options)

	//TODO: improvement
	var span opentracing.Span
	if httpClient.cfg.TurnOffJaeger || httpClient.jaegerTracer != nil && httpClient.jaegerTracer.IsEnabled() {
		span = httpClient.jaegerTracer.HttpClientTracer(ctx, req, operationName)
		defer span.Finish()
	}

	resp, err = httpClient.firstAttemptAndRetry(ctx, &retryConfig, req, operationName, options)
	if err != nil {
		return resp, err
	}

	//TODO: improvement
	if httpClient.jaegerTracer != nil && httpClient.jaegerTracer.IsEnabled() && span != nil {
		span.SetTag("http.response.status", resp.StatusCode)
		for k, v := range resp.Header {
			span.SetTag(fmt.Sprintf("http.response.header.%s", k), v)
		}
	}

	httpClient.logger.InfoKVf(ctx, logging.KV{"URL": url, "Status": resp.Status, "Headers": resp.Header}, "[%s] Received response", httpPatchMethod)
	return resp, nil
}

func (httpClient *httpClient) DELETE(ctx context.Context, url string, body io.Reader, opts ...option.Option) (*http.Response, error) {
	options := option.NewOptions(opts...)
	retryConfig := httpClient.getRetrySetting(ctx, httpDeleteMethod, url)
	operationName := httpClient.getOpNameFromOption(url, httpDeleteMethod, options)

	var resp *http.Response
	req, err := http.NewRequest(httpDeleteMethod, url, body)
	if err != nil {
		return resp, err
	}

	req.Header.Set("Content-Type", applicationJSON)
	httpClient.setHeaderFromOption(req, options)

	//TODO: improvement
	var span opentracing.Span
	if httpClient.cfg.TurnOffJaeger || httpClient.jaegerTracer != nil && httpClient.jaegerTracer.IsEnabled() {
		span = httpClient.jaegerTracer.HttpClientTracer(ctx, req, operationName)
		defer span.Finish()
	}

	resp, err = httpClient.firstAttemptAndRetry(ctx, &retryConfig, req, operationName, options)
	if err != nil {
		return resp, err
	}

	//TODO: improvement
	if httpClient.jaegerTracer != nil && httpClient.jaegerTracer.IsEnabled() && span != nil {
		span.SetTag("http.response.status", resp.StatusCode)
		for k, v := range resp.Header {
			span.SetTag(fmt.Sprintf("http.response.header.%s", k), v)
		}
	}

	httpClient.logger.InfoKVf(ctx, logging.KV{"URL": url, "Status": resp.Status, "Headers": resp.Header}, "[%s] Received response", httpPostMethod)
	return resp, nil
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
	var (
		bodyReader *bytes.Reader
		err        error
		resp       *http.Response
	)

	if req.Body != nil {
		reqData, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(reqData)
		req.Body = ioutil.NopCloser(bodyReader) // prevents closing the body between retries
	}

	// without retry
	if !retryConfig.Enabled {
		if httpClient.cfg.TurnOffCircuitBreaker || httpClient.circuitBreaker == nil || !httpClient.circuitBreaker.IsEnabled() {
			// no retry and circuit breaker
			return httpClient.sendHttpRequest(ctx, req, operationName, options)
		}

		// no retry with circuit breaker
		httpClient.circuitBreaker.Do(ctx, operationName,
			func() error {
				resp, err := httpClient.sendHttpRequest(ctx, req, operationName, options)
				if bodyReader != nil {
					_, _ = bodyReader.Seek(0, 0)
				}

				if err != nil {
					return err
				}

				if resp.StatusCode >= http.StatusInternalServerError {
					return err5xx
				}
				return nil
			},
			func(err error) error {
				return nil
			})
		return resp, err
	}

	// with retry
	for count := uint(0); count <= retryConfig.MaxRetryAttempts; count++ {
		if httpClient.cfg.TurnOffCircuitBreaker || httpClient.circuitBreaker == nil || !httpClient.circuitBreaker.IsEnabled() {
			// retry without circuit breaker
			resp, err := httpClient.sendHttpRequest(ctx, req, operationName, options)
			if bodyReader != nil {
				_, _ = bodyReader.Seek(0, 0)
			}

			if err == nil && resp.StatusCode >= http.StatusInternalServerError {
				err = err5xx
			}
			if err != nil {
				backOffDuration := defaultBackOffDuration
				if uint(len(retryConfig.BackOffDurations)) >= count {
					backOffDuration = retryConfig.BackOffDurations[count]
				}
				time.Sleep(backOffDuration)
				continue
			}
		} else {
			// retry with circuit beaker
			httpClient.circuitBreaker.Do(ctx, operationName,
				func() error {
					resp, err := httpClient.sendHttpRequest(ctx, req, operationName, options)
					if bodyReader != nil {
						_, _ = bodyReader.Seek(0, 0)
					}

					if err != nil {
						return err
					}

					if resp.StatusCode >= http.StatusInternalServerError {
						return err5xx
					}
					return nil
				},
				func(err error) error {
					return nil
				})

			if err != nil {
				backOffDuration := defaultBackOffDuration
				if uint(len(retryConfig.BackOffDurations)) >= count {
					backOffDuration = retryConfig.BackOffDurations[count]
				}
				time.Sleep(backOffDuration)
				continue
			}
		}
		break
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

	if httpClient.cfg.TurnOffNewrelic || httpClient.newrelicTracer == nil || !httpClient.newrelicTracer.IsEnabled() {
		return client.Do(req)
	}

	es, err := httpClient.newrelicTracer.RecordExternalMetric(req, name)
	if err == nil {
		defer es.End()
	}

	response, err := client.Do(req)
	return response, err
}

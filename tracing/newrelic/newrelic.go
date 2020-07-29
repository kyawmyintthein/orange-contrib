package newrelic

import (
	"net/http"

	newrelic "github.com/newrelic/go-agent"
)

type NewrelicTracer interface {
	IsEnabled() bool
	RecordExternalMetric(*http.Request, string) (*newrelic.ExternalSegment, error)
}

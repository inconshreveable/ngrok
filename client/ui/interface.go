package ui

import (
	metrics "github.com/rcrowley/go-metrics"
	"net/http"
)

type HttpRequest interface {
	GetRequest() *http.Request
	GetResponse() *http.Response
}

type State interface {
	GetVersion() string
	GetPublicUrl() string
	GetLocalAddr() string
	GetStatus() string
	GetHistory() []HttpRequest
	IsStopping() bool
	GetConnectionMetrics() (metrics.Meter, metrics.Timer)
	GetRequestMetrics() (metrics.Meter, metrics.Timer)
	GetBytesInMetrics() (metrics.Counter, metrics.Histogram)
	GetBytesOutMetrics() (metrics.Counter, metrics.Histogram)
}

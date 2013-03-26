package ui

import (
	metrics "github.com/rcrowley/go-metrics"
	"ngrok/proto"
)

type State interface {
	GetVersion() string
	GetPublicUrl() string
	GetLocalAddr() string
	GetStatus() string
	GetProtocol() proto.Protocol
	IsStopping() bool
	GetConnectionMetrics() (metrics.Meter, metrics.Timer)
	GetBytesInMetrics() (metrics.Counter, metrics.Histogram)
	GetBytesOutMetrics() (metrics.Counter, metrics.Histogram)
}

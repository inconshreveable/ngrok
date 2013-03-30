package ui

import (
	metrics "github.com/inconshreveable/go-metrics"
	"ngrok/proto"
)

type State interface {
	GetVersion() string
	GetPublicUrl() string
	GetLocalAddr() string
	GetStatus() string
	GetProtocol() proto.Protocol
	GetWebPort() int
	IsStopping() bool
	GetConnectionMetrics() (metrics.Meter, metrics.Timer)
	GetBytesInMetrics() (metrics.Counter, metrics.Histogram)
	GetBytesOutMetrics() (metrics.Counter, metrics.Histogram)
}

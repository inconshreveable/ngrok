package ui

import (
	metrics "github.com/inconshreveable/go-metrics"
	"ngrok/proto"
)

type State interface {
	GetClientVersion() string
	GetServerVersion() string
	GetNewVersion() string
	GetPublicUrl() string
	GetLocalAddr() string
	GetStatus() string
	GetProtocol() proto.Protocol
	GetWebPort() int
	GetConnectionMetrics() (metrics.Meter, metrics.Timer)
	GetBytesInMetrics() (metrics.Counter, metrics.Histogram)
	GetBytesOutMetrics() (metrics.Counter, metrics.Histogram)
}

package ui

import (
	metrics "github.com/inconshreveable/go-metrics"
	"github.com/inconshreveable/ngrok/proto"
)

type UpdateStatus int

const (
	UpdateNone = -1 * iota
	UpdateInstalling
	UpdateReady
	UpdateError
)

type State interface {
	GetClientVersion() string
	GetServerVersion() string
	GetUpdate() UpdateStatus
	GetPublicUrl() string
	GetLocalAddr() string
	GetStatus() string
	GetProtocol() proto.Protocol
	GetWebPort() int
	GetConnectionMetrics() (metrics.Meter, metrics.Timer)
	GetBytesInMetrics() (metrics.Counter, metrics.Histogram)
	GetBytesOutMetrics() (metrics.Counter, metrics.Histogram)
}

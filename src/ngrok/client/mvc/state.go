package mvc

import (
	metrics "github.com/inconshreveable/go-metrics"
	"ngrok/proto"
)

type UpdateStatus int

const (
	UpdateNone = -1 * iota
	UpdateInstalling
	UpdateReady
	UpdateError
)

type ConnStatus int

const (
	ConnConnecting = iota
	ConnReconnecting
	ConnOnline
)

type Tunnel struct {
	PublicUrl string
	Protocol proto.Protocol
	LocalAddr string
}

type State interface {
	GetClientVersion() string
	GetServerVersion() string
	GetUpdate() UpdateStatus
	GetTunnels() []Tunnel
	GetStatus() string
	GetWebPort() int
	GetConnectionMetrics() (metrics.Meter, metrics.Timer)
	GetBytesInMetrics() (metrics.Counter, metrics.Histogram)
	GetBytesOutMetrics() (metrics.Counter, metrics.Histogram)
}

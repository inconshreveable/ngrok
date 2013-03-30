package client

import (
	metrics "github.com/inconshreveable/go-metrics"
	"ngrok/proto"
)

// client state
type State struct {
	id        string
	publicUrl string
	protocol  proto.Protocol
	opts      *Options
	metrics   *ClientMetrics

	// just for UI purposes
	status   string
	stopping bool
}

// implement client.ui.State
func (s State) GetVersion() string          { return "" }
func (s State) GetPublicUrl() string        { return s.publicUrl }
func (s State) GetLocalAddr() string        { return s.opts.localaddr }
func (s State) GetWebPort() int             { return s.opts.webport }
func (s State) GetStatus() string           { return s.status }
func (s State) GetProtocol() proto.Protocol { return s.protocol }
func (s State) IsStopping() bool            { return s.stopping }

func (s State) GetConnectionMetrics() (metrics.Meter, metrics.Timer) {
	return s.metrics.connMeter, s.metrics.connTimer
}

func (s State) GetBytesInMetrics() (metrics.Counter, metrics.Histogram) {
	return s.metrics.bytesInCount, s.metrics.bytesIn
}

func (s State) GetBytesOutMetrics() (metrics.Counter, metrics.Histogram) {
	return s.metrics.bytesOutCount, s.metrics.bytesOut
}

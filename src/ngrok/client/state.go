package client

import (
	"fmt"
	metrics "github.com/inconshreveable/go-metrics"
	"ngrok/client/ui"
	"ngrok/proto"
	"ngrok/version"
)

// client state
type State struct {
	id            string
	publicUrl     string
	publicPort    int
	serverVersion string
	update        ui.UpdateStatus
	protocol      proto.Protocol
	opts          *Options
	metrics       *ClientMetrics

	// just for UI purposes
	status string
}

// implement client.ui.State
func (s State) GetClientVersion() string    { return version.MajorMinor() }
func (s State) GetServerVersion() string    { return s.serverVersion }
func (s State) GetLocalAddr() string        { return s.opts.localaddr }
func (s State) GetWebPort() int             { return s.opts.webport }
func (s State) GetStatus() string           { return s.status }
func (s State) GetProtocol() proto.Protocol { return s.protocol }
func (s State) GetUpdate() ui.UpdateStatus  { return s.update }

func (s State) GetPublicUrl() string {
	publicUrl := s.publicUrl
	if s.publicPort != 80 && s.publicPort != 443 {
		publicUrl += fmt.Sprintf(":%d", s.publicPort)
	}
	return publicUrl
}

func (s State) GetConnectionMetrics() (metrics.Meter, metrics.Timer) {
	return s.metrics.connMeter, s.metrics.connTimer
}

func (s State) GetBytesInMetrics() (metrics.Counter, metrics.Histogram) {
	return s.metrics.bytesInCount, s.metrics.bytesIn
}

func (s State) GetBytesOutMetrics() (metrics.Counter, metrics.Histogram) {
	return s.metrics.bytesOutCount, s.metrics.bytesOut
}

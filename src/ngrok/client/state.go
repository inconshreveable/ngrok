package client

import (
	metrics "github.com/rcrowley/go-metrics"
	"ngrok/client/ui"
)

// client state
type State struct {
	id        string
	ui        *ui.Ui
	publicUrl string
	protocol  string
	history   *RequestHistory
	opts      *Options
	metrics   *ClientMetrics

	// just for UI purposes
	status         string
	historyEntries []*RequestHistoryEntry
	stopping       bool
}

// implement client.ui.State
func (s State) GetVersion() string   { return "" }
func (s State) GetPublicUrl() string { return s.publicUrl }
func (s State) GetLocalAddr() string { return s.opts.localaddr }
func (s State) GetStatus() string    { return s.status }
func (s State) GetProtocol() string  { return s.protocol }
func (s State) GetHistory() []ui.HttpRequest {
	// go sucks
	historyEntries := make([]ui.HttpRequest, len(s.historyEntries))

	for i, entry := range s.historyEntries {
		historyEntries[i] = entry
	}
	return historyEntries
}
func (s State) IsStopping() bool { return s.stopping }

func (s State) GetConnectionMetrics() (metrics.Meter, metrics.Timer) {
	return s.metrics.connMeter, s.metrics.connTimer
}

func (s State) GetRequestMetrics() (metrics.Meter, metrics.Timer) {
	return s.metrics.reqMeter, s.metrics.reqTimer
}

func (s State) GetBytesInMetrics() (metrics.Counter, metrics.Histogram) {
	return s.metrics.bytesInCount, s.metrics.bytesIn
}

func (s State) GetBytesOutMetrics() (metrics.Counter, metrics.Histogram) {
	return s.metrics.bytesOutCount, s.metrics.bytesOut
}

func (s *State) Update() {
	s.ui.Updates.In() <- *s
}

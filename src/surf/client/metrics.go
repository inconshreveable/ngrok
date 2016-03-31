package client

import (
	metrics "github.com/rcrowley/go-metrics"
)

const (
	sampleSize  int     = 1028
	sampleAlpha float64 = 0.015
)

type ClientMetrics struct {
	// metrics
	connGauge       metrics.Gauge
	connMeter       metrics.Meter
	connTimer       metrics.Timer
	proxySetupTimer metrics.Timer
	bytesIn         metrics.Histogram
	bytesOut        metrics.Histogram
	bytesInCount    metrics.Counter
	bytesOutCount   metrics.Counter
}

func NewClientMetrics() *ClientMetrics {
	return &ClientMetrics{
		connGauge:       metrics.NewGauge(),
		connMeter:       metrics.NewMeter(),
		connTimer:       metrics.NewTimer(),
		proxySetupTimer: metrics.NewTimer(),
		bytesIn:         metrics.NewHistogram(metrics.NewExpDecaySample(sampleSize, sampleAlpha)),
		bytesOut:        metrics.NewHistogram(metrics.NewExpDecaySample(sampleSize, sampleAlpha)),
		bytesInCount:    metrics.NewCounter(),
		bytesOutCount:   metrics.NewCounter(),
	}
}

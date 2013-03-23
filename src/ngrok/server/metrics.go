package server

import (
	log "code.google.com/p/log4go"
	"encoding/json"
	gometrics "github.com/rcrowley/go-metrics"
	"time"
)

var reportInterval = 30 * time.Second

var metrics struct {
	windowsCounter gometrics.Counter
	linuxCounter   gometrics.Counter
	osxCounter     gometrics.Counter
	otherCounter   gometrics.Counter
	/*
	   bytesInCount gometrics.Counter
	   bytesOutCount gometrics.Counter
	*/

	/*
	   tunnelGauge gometrics.Gauge
	   tcpTunnelGauge gometrics.Gauge
	   requestGauge gometrics.Gauge
	*/

	tunnelMeter        gometrics.Meter
	tcpTunnelMeter     gometrics.Meter
	requestMeter       gometrics.Meter
	lostHeartbeatMeter gometrics.Meter

	requestTimer gometrics.Timer
}

func init() {
	metrics.windowsCounter = gometrics.NewCounter()
	metrics.linuxCounter = gometrics.NewCounter()
	metrics.osxCounter = gometrics.NewCounter()
	metrics.otherCounter = gometrics.NewCounter()
	/*
	   metrics.bytesInCount = gometrics.NewCounter()
	   metrics.bytesOutCount = gometrics.NewCounter()
	*/

	/*
	   metrics.tunnelGauge = gometrics.NewGauge()
	   metrics.tcpTunnelGauge = gometrics.NewGauge()
	   metrics.requestGauge = gometrics.NewGauge()
	*/

	metrics.tunnelMeter = gometrics.NewMeter()
	metrics.tcpTunnelMeter = gometrics.NewMeter()
	metrics.requestMeter = gometrics.NewMeter()
	metrics.lostHeartbeatMeter = gometrics.NewMeter()

	metrics.requestTimer = gometrics.NewTimer()

	go func() {
		time.Sleep(reportInterval)
		log.Info("Server metrics: %s", MetricsJson())
	}()
}

func MetricsJson() []byte {
	buffer, _ := json.Marshal(map[string]interface{}{
		"windows":            metrics.windowsCounter.Count(),
		"linux":              metrics.linuxCounter.Count(),
		"osx":                metrics.osxCounter.Count(),
		"other":              metrics.otherCounter.Count(),
		"tunnelMeter.count":  metrics.tunnelMeter.Count(),
		"tunnelMeter.m1":     metrics.tunnelMeter.Rate1(),
		"requestMeter.count": metrics.requestMeter.Count(),
		"requestMeter.m1":    metrics.requestMeter.Rate1(),
	})
	return buffer
}

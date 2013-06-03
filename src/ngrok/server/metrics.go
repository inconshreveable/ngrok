package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	gometrics "github.com/inconshreveable/go-metrics"
	"io/ioutil"
	"net/http"
	"ngrok/conn"
	"ngrok/log"
	"os"
	"time"
)

var metrics Metrics

func init() {
	keenApiKey := os.Getenv("KEEN_API_KEY")

	if keenApiKey != "" {
		metrics = NewKeenIoMetrics()
	} else {
		metrics = NewLocalMetrics(30 * time.Second)
	}

	metrics.AddLogPrefix("metrics")
}

type Metrics interface {
	log.Logger
	OpenConnection(*Tunnel, conn.Conn)
	CloseConnection(*Tunnel, conn.Conn, time.Time, int64, int64)
	OpenTunnel(*Tunnel)
	CloseTunnel(*Tunnel)
}

type LocalMetrics struct {
	log.Logger
	reportInterval time.Duration
	windowsCounter gometrics.Counter
	linuxCounter   gometrics.Counter
	osxCounter     gometrics.Counter
	otherCounter   gometrics.Counter

	tunnelMeter        gometrics.Meter
	tcpTunnelMeter     gometrics.Meter
	httpTunnelMeter    gometrics.Meter
	connMeter          gometrics.Meter
	lostHeartbeatMeter gometrics.Meter

	connTimer gometrics.Timer

	bytesInCount  gometrics.Counter
	bytesOutCount gometrics.Counter

	/*
	   tunnelGauge gometrics.Gauge
	   tcpTunnelGauge gometrics.Gauge
	   connGauge gometrics.Gauge
	*/
}

func NewLocalMetrics(reportInterval time.Duration) *LocalMetrics {
	metrics := LocalMetrics{
		Logger:         log.NewPrefixLogger(),
		reportInterval: reportInterval,
		windowsCounter: gometrics.NewCounter(),
		linuxCounter:   gometrics.NewCounter(),
		osxCounter:     gometrics.NewCounter(),
		otherCounter:   gometrics.NewCounter(),

		tunnelMeter:        gometrics.NewMeter(),
		tcpTunnelMeter:     gometrics.NewMeter(),
		httpTunnelMeter:    gometrics.NewMeter(),
		connMeter:          gometrics.NewMeter(),
		lostHeartbeatMeter: gometrics.NewMeter(),

		connTimer: gometrics.NewTimer(),

		bytesInCount:  gometrics.NewCounter(),
		bytesOutCount: gometrics.NewCounter(),

		/*
		   metrics.tunnelGauge = gometrics.NewGauge(),
		   metrics.tcpTunnelGauge = gometrics.NewGauge(),
		   metrics.connGauge = gometrics.NewGauge(),
		*/
	}

	go metrics.Report()

	return &metrics
}

func (m *LocalMetrics) OpenTunnel(t *Tunnel) {
	m.tunnelMeter.Mark(1)

	switch t.regMsg.OS {
	case "windows":
		m.windowsCounter.Inc(1)
	case "linux":
		m.linuxCounter.Inc(1)
	case "darwin":
		m.osxCounter.Inc(1)
	default:
		m.otherCounter.Inc(1)
	}

	switch t.regMsg.Protocol {
	case "tcp":
		m.tcpTunnelMeter.Mark(1)
	case "http":
		m.httpTunnelMeter.Mark(1)
	}
}

func (m *LocalMetrics) CloseTunnel(t *Tunnel) {
}

func (m *LocalMetrics) OpenConnection(t *Tunnel, c conn.Conn) {
	m.connMeter.Mark(1)
}

func (m *LocalMetrics) CloseConnection(t *Tunnel, c conn.Conn, start time.Time, bytesIn, bytesOut int64) {
	m.bytesInCount.Inc(bytesIn)
	m.bytesOutCount.Inc(bytesOut)
}

func (m *LocalMetrics) Report() {
	m.Info("Reporting every %d seconds", int(m.reportInterval.Seconds()))

	for {
		time.Sleep(m.reportInterval)
		buffer, err := json.Marshal(map[string]interface{}{
			"windows":               m.windowsCounter.Count(),
			"linux":                 m.linuxCounter.Count(),
			"osx":                   m.osxCounter.Count(),
			"other":                 m.otherCounter.Count(),
			"httpTunnelMeter.count": m.httpTunnelMeter.Count(),
			"tcpTunnelMeter.count":  m.tcpTunnelMeter.Count(),
			"tunnelMeter.count":     m.tunnelMeter.Count(),
			"tunnelMeter.m1":        m.tunnelMeter.Rate1(),
			"connMeter.count":       m.connMeter.Count(),
			"connMeter.m1":          m.connMeter.Rate1(),
			"bytesIn.count":         m.bytesInCount.Count(),
			"bytesOut.count":        m.bytesOutCount.Count(),
		})

		if err != nil {
			m.Error("Failed to serialize metrics: %v", err)
			continue
		}

		m.Info("Reporting: %s", buffer)
	}
}

type KeenIoRequest struct {
	Path string
	Body []byte
}

type KeenIoMetrics struct {
	log.Logger
	ApiKey       string
	ProjectToken string
	HttpClient   http.Client
	Requests     chan *KeenIoRequest
}

func NewKeenIoMetrics() *KeenIoMetrics {
	k := &KeenIoMetrics{
		Logger:       log.NewPrefixLogger(),
		ApiKey:       os.Getenv("KEEN_API_KEY"),
		ProjectToken: os.Getenv("KEEN_PROJECT_TOKEN"),
		Requests:     make(chan *KeenIoRequest, 100),
	}

	go func() {
		for req := range k.Requests {
			k.AuthedRequest("POST", req.Path, bytes.NewReader(req.Body))
		}
	}()

	return k
}

func (k *KeenIoMetrics) AuthedRequest(method, path string, body *bytes.Reader) (resp *http.Response, err error) {
	path = fmt.Sprintf("https://api.keen.io/3.0/projects/%s%s", k.ProjectToken, path)
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return
	}

	req.Header.Add("Authorization", k.ApiKey)

	if body != nil {
		req.Header.Add("Content-Type", "application/json")
		req.ContentLength = int64(body.Len())
	}

	resp, err = k.HttpClient.Do(req)

	if err != nil {
		k.Error("Failed to send metric event to keen.io %v", err)
	} else {
		defer resp.Body.Close()
		if resp.StatusCode != 201 {
			bytes, _ := ioutil.ReadAll(resp.Body)
			k.Error("Got %v response from keen.io: %s", resp.StatusCode, bytes)
		}
	}

	return
}

func (k *KeenIoMetrics) OpenConnection(t *Tunnel, c conn.Conn) {
}

func (k *KeenIoMetrics) CloseConnection(t *Tunnel, c conn.Conn, start time.Time, in, out int64) {
	buf, err := json.Marshal(struct {
		Keen               KeenStruct `json:"keen"`
		OS                 string
		ClientId           string
		Protocol           string
		Url                string
		User               string
		Version            string
		Reason             string
		HttpAuth           bool
		Subdomain          bool
		TunnelDuration     float64
		ConnectionDuration float64
		BytesIn            int64
		BytesOut           int64
	}{
		Keen: KeenStruct{
			Timestamp: start.UTC().Format("2006-01-02T15:04:05.000Z"),
		},
		OS:                 t.regMsg.OS,
		ClientId:           t.regMsg.ClientId,
		Protocol:           t.regMsg.Protocol,
		Url:                t.url,
		User:               t.regMsg.User,
		Version:            t.regMsg.MmVersion,
		HttpAuth:           t.regMsg.HttpAuth != "",
		Subdomain:          t.regMsg.Subdomain != "",
		TunnelDuration:     time.Since(t.start).Seconds(),
		ConnectionDuration: time.Since(start).Seconds(),
		BytesIn:            in,
		BytesOut:           out,
	})

	if err != nil {
		k.Error("Error serializing metric %v", err)
	} else {
		k.Requests <- &KeenIoRequest{Path: "/events/CloseConnection", Body: buf}
	}
}

func (k *KeenIoMetrics) OpenTunnel(t *Tunnel) {
}

type KeenStruct struct {
	Timestamp string `json:"timestamp"`
}

func (k *KeenIoMetrics) CloseTunnel(t *Tunnel) {
	buf, err := json.Marshal(struct {
		Keen      KeenStruct `json:"keen"`
		OS        string
		ClientId  string
		Protocol  string
		Url       string
		User      string
		Version   string
		Reason    string
		Duration  float64
		HttpAuth  bool
		Subdomain bool
	}{
		Keen: KeenStruct{
			Timestamp: t.start.UTC().Format("2006-01-02T15:04:05.000Z"),
		},
		OS:       t.regMsg.OS,
		ClientId: t.regMsg.ClientId,
		Protocol: t.regMsg.Protocol,
		Url:      t.url,
		User:     t.regMsg.User,
		Version:  t.regMsg.MmVersion,
		//Reason: reason,
		Duration:  time.Since(t.start).Seconds(),
		HttpAuth:  t.regMsg.HttpAuth != "",
		Subdomain: t.regMsg.Subdomain != "",
	})

	if err != nil {
		k.Error("Error serializing metric %v", err)
		return
	} else {
		k.Requests <- &KeenIoRequest{Path: "/events/CloseTunnel", Body: buf}
	}
}

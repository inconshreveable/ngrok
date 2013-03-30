package proto

import (
	metrics "github.com/inconshreveable/go-metrics"
	"net/http"
	"net/http/httputil"
	"ngrok/conn"
	"ngrok/util"
	"time"
)

type HttpTxn struct {
	Req      *http.Request
	Resp     *http.Response
	Start    time.Time
	Duration time.Duration
}

type Http struct {
	Txns     *util.Broadcast
	reqGauge metrics.Gauge
	reqMeter metrics.Meter
	reqTimer metrics.Timer
}

func NewHttp() *Http {
	return &Http{
		Txns:     util.NewBroadcast(),
		reqGauge: metrics.NewGauge(),
		reqMeter: metrics.NewMeter(),
		reqTimer: metrics.NewTimer(),
	}
}

func (h *Http) GetName() string { return "http" }

func (h *Http) WrapConn(c conn.Conn) conn.Conn {
	tee := conn.NewTee(c)
	lastTxn := make(chan *HttpTxn)
	go h.readRequests(tee, lastTxn)
	go h.readResponses(tee, lastTxn)
	return tee
}

func (h *Http) readRequests(tee *conn.Tee, lastTxn chan *HttpTxn) {
	for {
		req, err := http.ReadRequest(tee.WriteBuffer())
		if err != nil {
			// no more requests to be read, we're done
			break
		}

		// make sure we read the body of the request so that
		// we don't block the writer 
		_, _ = httputil.DumpRequest(req, true)

		h.reqMeter.Mark(1)
		txn := &HttpTxn{Req: req, Start: time.Now()}
		lastTxn <- txn
		h.Txns.In() <- txn
	}
}

func (h *Http) readResponses(tee *conn.Tee, lastTxn chan *HttpTxn) {
	for {
		var err error
		txn := <-lastTxn
		txn.Resp, err = http.ReadResponse(tee.ReadBuffer(), txn.Req)
		txn.Duration = time.Since(txn.Start)
		h.reqTimer.Update(txn.Duration)
		if err != nil {
			tee.Warn("Error reading response from server: %v", err)
			// no more responses to be read, we're done
			break
		}
		// make sure we read the body of the response so that
		// we don't block the reader 
		_, _ = httputil.DumpResponse(txn.Resp, true)

		h.Txns.In() <- txn
	}
}

package proto

import (
	"bytes"
	metrics "github.com/inconshreveable/go-metrics"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"ngrok/conn"
	"ngrok/util"
	"time"
)

type HttpRequest struct {
	*http.Request
	BodyBytes []byte
}

type HttpResponse struct {
	*http.Response
	BodyBytes []byte
}

type HttpTxn struct {
	Req      *HttpRequest
	Resp     *HttpResponse
	Start    time.Time
	Duration time.Duration
	UserData interface{}
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

func extractBody(r io.Reader) ([]byte, io.ReadCloser, error) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r)
	return buf.Bytes(), ioutil.NopCloser(buf), err
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
		_, err = httputil.DumpRequest(req, true)

		h.reqMeter.Mark(1)
		if err != nil {
			tee.Warn("Failed to extract request body: %v", err)
		}

		txn := &HttpTxn{Start: time.Now()}
		txn.Req = &HttpRequest{Request: req}
		txn.Req.BodyBytes, txn.Req.Body, err = extractBody(req.Body)

		lastTxn <- txn
		h.Txns.In() <- txn
	}
}

func (h *Http) readResponses(tee *conn.Tee, lastTxn chan *HttpTxn) {
	for {
		var err error
		txn := <-lastTxn
		resp, err := http.ReadResponse(tee.ReadBuffer(), txn.Req.Request)
		txn.Duration = time.Since(txn.Start)
		h.reqTimer.Update(txn.Duration)
		if err != nil {
			tee.Warn("Error reading response from server: %v", err)
			// no more responses to be read, we're done
			break
		}
		// make sure we read the body of the response so that
		// we don't block the reader
		_, _ = httputil.DumpResponse(resp, true)

		txn.Resp = &HttpResponse{Response: resp}
		txn.Resp.BodyBytes, txn.Resp.Body, err = extractBody(resp.Body)
		if err != nil {
			tee.Warn("Failed to extract response body: %v", err)
		}

		h.Txns.In() <- txn
	}
}

package proto

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"ngrok/conn"
	"ngrok/util"
	"strings"
	"sync"
	"time"

	metrics "github.com/rcrowley/go-metrics"
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
	Req         *HttpRequest
	Resp        *HttpResponse
	Start       time.Time
	Duration    time.Duration
	UserCtx     interface{}
	ConnUserCtx interface{}
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

func (h *Http) WrapConn(c conn.Conn, ctx interface{}) conn.Conn {
	tee := conn.NewTee(c)
	lastTxn := make(chan *HttpTxn)
	go h.readRequests(tee, lastTxn, ctx)
	go h.readResponses(tee, lastTxn)
	return tee
}

func (h *Http) readRequests(tee *conn.Tee, lastTxn chan *HttpTxn, connCtx interface{}) {
	defer close(lastTxn)

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

		// golang's ReadRequest/DumpRequestOut is broken. Fix up the request so it works later
		req.URL.Scheme = "http"
		req.URL.Host = req.Host

		txn := &HttpTxn{Start: time.Now(), ConnUserCtx: connCtx}
		txn.Req = &HttpRequest{Request: req}
		if req.Body != nil {
			txn.Req.BodyBytes, txn.Req.Body, err = extractBody(req.Body)
			if err != nil {
				tee.Warn("Failed to extract request body: %v", err)
			}
		}

		lastTxn <- txn
		h.Txns.In() <- txn
	}
}

func (h *Http) readResponses(tee *conn.Tee, lastTxn chan *HttpTxn) {
	for txn := range lastTxn {
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
		// apparently, Body can be nil in some cases
		if resp.Body != nil {
			txn.Resp.BodyBytes, txn.Resp.Body, err = extractBody(resp.Body)
			if err != nil {
				tee.Warn("Failed to extract response body: %v", err)
			}
		}

		h.Txns.In() <- txn

		// XXX: remove web socket shim in favor of a real websocket protocol analyzer
		if txn.Req.Header.Get("Upgrade") == "websocket" {
			tee.Info("Upgrading to websocket")
			var wg sync.WaitGroup

			// shim for websockets
			// in order for websockets to work, we need to continue reading all of the
			// the bytes in the analyzer so that the joined connections will continue
			// sending bytes to each other
			wg.Add(2)
			go func() {
				ioutil.ReadAll(tee.WriteBuffer())
				wg.Done()
			}()

			go func() {
				ioutil.ReadAll(tee.ReadBuffer())
				wg.Done()
			}()

			wg.Wait()
			break
		}
	}
}

// we have to vendor DumpRequestOut because it's broken and the fix won't be in until at least 1.4
// XXX: remove this all in favor of actually parsing the HTTP traffic ourselves for more transparent
// replay and inspection, regardless of when it gets fixed in stdlib

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// One of the copies, say from b to r2, could be avoided by using a more
// elaborate trick where the other copy is made during Request/Response.Write.
// This would complicate things too much, given that these functions are for
// debugging only.
func drainBody(b io.ReadCloser) (r1, r2 io.ReadCloser, err error) {
	var buf bytes.Buffer
	if _, err = buf.ReadFrom(b); err != nil {
		return nil, nil, err
	}
	if err = b.Close(); err != nil {
		return nil, nil, err
	}
	return ioutil.NopCloser(&buf), ioutil.NopCloser(bytes.NewReader(buf.Bytes())), nil
}

// dumpConn is a net.Conn which writes to Writer and reads from Reader
type dumpConn struct {
	io.Writer
	io.Reader
}

func (c *dumpConn) Close() error                       { return nil }
func (c *dumpConn) LocalAddr() net.Addr                { return nil }
func (c *dumpConn) RemoteAddr() net.Addr               { return nil }
func (c *dumpConn) SetDeadline(t time.Time) error      { return nil }
func (c *dumpConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *dumpConn) SetWriteDeadline(t time.Time) error { return nil }

type neverEnding byte

func (b neverEnding) Read(p []byte) (n int, err error) {
	for i := range p {
		p[i] = byte(b)
	}
	return len(p), nil
}

// DumpRequestOut is like DumpRequest but includes
// headers that the standard http.Transport adds,
// such as User-Agent.
func DumpRequestOut(req *http.Request, body bool) ([]byte, error) {
	save := req.Body
	dummyBody := false
	if !body || req.Body == nil {
		req.Body = nil
		if req.ContentLength != 0 {
			req.Body = ioutil.NopCloser(io.LimitReader(neverEnding('x'), req.ContentLength))
			dummyBody = true
		}
	} else {
		var err error
		save, req.Body, err = drainBody(req.Body)
		if err != nil {
			return nil, err
		}
	}

	// Since we're using the actual Transport code to write the request,
	// switch to http so the Transport doesn't try to do an SSL
	// negotiation with our dumpConn and its bytes.Buffer & pipe.
	// The wire format for https and http are the same, anyway.
	reqSend := req
	if req.URL.Scheme == "https" {
		reqSend = new(http.Request)
		*reqSend = *req
		reqSend.URL = new(url.URL)
		*reqSend.URL = *req.URL
		reqSend.URL.Scheme = "http"
	}

	// Use the actual Transport code to record what we would send
	// on the wire, but not using TCP.  Use a Transport with a
	// custom dialer that returns a fake net.Conn that waits
	// for the full input (and recording it), and then responds
	// with a dummy response.
	var buf bytes.Buffer // records the output
	pr, pw := io.Pipe()
	dr := &delegateReader{c: make(chan io.Reader)}
	// Wait for the request before replying with a dummy response:
	go func() {
		req, _ := http.ReadRequest(bufio.NewReader(pr))
		// THIS IS THE PART THAT'S BROKEN IN THE STDLIB (as of Go 1.3)
		if req != nil && req.Body != nil {
			ioutil.ReadAll(req.Body)
		}
		dr.c <- strings.NewReader("HTTP/1.1 204 No Content\r\n\r\n")
	}()

	t := &http.Transport{
		Dial: func(net, addr string) (net.Conn, error) {
			return &dumpConn{io.MultiWriter(&buf, pw), dr}, nil
		},
	}
	defer t.CloseIdleConnections()

	_, err := t.RoundTrip(reqSend)

	req.Body = save
	if err != nil {
		return nil, err
	}
	dump := buf.Bytes()

	// If we used a dummy body above, remove it now.
	// TODO: if the req.ContentLength is large, we allocate memory
	// unnecessarily just to slice it off here.  But this is just
	// a debug function, so this is acceptable for now. We could
	// discard the body earlier if this matters.
	if dummyBody {
		if i := bytes.Index(dump, []byte("\r\n\r\n")); i >= 0 {
			dump = dump[:i+4]
		}
	}
	return dump, nil
}

// delegateReader is a reader that delegates to another reader,
// once it arrives on a channel.
type delegateReader struct {
	c chan io.Reader
	r io.Reader // nil until received from c
}

func (r *delegateReader) Read(p []byte) (int, error) {
	if r.r == nil {
		r.r = <-r.c
	}
	return r.r.Read(p)
}

// Return value if nonempty, def otherwise.
func valueOrDefault(value, def string) string {
	if value != "" {
		return value
	}
	return def
}

var reqWriteExcludeHeaderDump = map[string]bool{
	"Host":              true, // not in Header map anyway
	"Content-Length":    true,
	"Transfer-Encoding": true,
	"Trailer":           true,
}

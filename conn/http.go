package conn

import (
	"net/http"
	"net/http/httputil"
)

func ParseHttp(tee *Tee, reqs chan *http.Request, resps chan *http.Response) {
	lastReq := make(chan *http.Request)

	go func() {
		for {
			req, err := http.ReadRequest(tee.ReadBuffer())
			if err != nil {
				// no more requests to be read, we're done
				break
			}
			lastReq <- req
			// make sure we read the body of the request so that
			// we don't block the reader 
			_, _ = httputil.DumpRequest(req, true)
			reqs <- req
		}
	}()

	go func() {
		for {
			req := <-lastReq
			resp, err := http.ReadResponse(tee.WriteBuffer(), req)
			if err != nil {
				tee.Warn("Error reading response from server: %v", err)
				// no more responses to be read, we're done
				break
			}
			// make sure we read the body of the response so that
			// we don't block the writer 
			_, _ = httputil.DumpResponse(resp, true)
			resps <- resp
		}
	}()

}

package client

import (
	"container/list"
	"net/http"
	"time"
)

type RequestHistoryEntry struct {
	req      *http.Request
	resp     *http.Response
	start    time.Time
	duration time.Duration
}

type RequestHistory struct {
	maxSize    int
	reqToEntry map[*http.Request]*RequestHistoryEntry
	reqs       chan *http.Request
	resps      chan *http.Response
	history    *list.List
	onChange   func([]*RequestHistoryEntry)
	metrics    *ClientMetrics
}

func NewRequestHistory(maxSize int, metrics *ClientMetrics, onChange func([]*RequestHistoryEntry)) *RequestHistory {
	rh := &RequestHistory{
		maxSize:    maxSize,
		reqToEntry: make(map[*http.Request]*RequestHistoryEntry),
		reqs:       make(chan *http.Request),
		resps:      make(chan *http.Response),
		history:    list.New(),
		onChange:   onChange,
		metrics:    metrics,
	}

	go func() {
		for {
			select {
			case req := <-rh.reqs:
				rh.addRequest(req)

			case resp := <-rh.resps:
				rh.addResponse(resp)
			}
		}
	}()

	return rh
}

func (rh *RequestHistory) addRequest(req *http.Request) {
	rh.metrics.reqMeter.Mark(1)
	if rh.history.Len() >= rh.maxSize {
		entry := rh.history.Remove(rh.history.Back()).(*RequestHistoryEntry)
		delete(rh.reqToEntry, entry.req)
	}

	entry := &RequestHistoryEntry{req: req, start: time.Now()}
	rh.reqToEntry[req] = entry
	rh.history.PushFront(entry)
	rh.onChange(rh.copy())
}

func (rh *RequestHistory) addResponse(resp *http.Response) {
	if entry, ok := rh.reqToEntry[resp.Request]; ok {
		entry.duration = time.Since(entry.start)
		rh.metrics.reqTimer.Update(entry.duration)

		entry.resp = resp
		rh.onChange(rh.copy())
	} else {
		// XXX: log warning instead of panic
		panic("no request for response!")
	}
}

func (rh *RequestHistory) copy() []*RequestHistoryEntry {
	entries := make([]*RequestHistoryEntry, rh.history.Len())
	i := 0
	for e := rh.history.Front(); e != nil; e = e.Next() {
		// force a copy
		entry := *(e.Value.(*RequestHistoryEntry))
		entries[i] = &entry
		i++
	}
	return entries
}

func (rhe *RequestHistoryEntry) GetRequest() *http.Request {
	return rhe.req
}

func (rhe *RequestHistoryEntry) GetResponse() *http.Response {
	return rhe.resp
}

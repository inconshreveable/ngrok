package term

import (
	termbox "github.com/nsf/termbox-go"
	"ngrok/log"
	"ngrok/proto"
	"ngrok/util"
)

const (
	size = 10
)

type HttpView struct {
	httpProto    *proto.Http
	HttpRequests *util.Ring
	shutdown     chan int
	flush        chan int
	*area
	log.Logger
}

func colorFor(status string) termbox.Attribute {
	switch status[0] {
	case '3':
		return termbox.ColorCyan
	case '4':
		return termbox.ColorYellow
	case '5':
		return termbox.ColorRed
	default:
	}
	return termbox.ColorWhite
}

func NewHttp(proto *proto.Http, flush, shutdown chan int, x, y int) *HttpView {
	v := &HttpView{
		httpProto:    proto,
		HttpRequests: util.NewRing(size),
		area:         NewArea(x, y, 70, size+5),
		shutdown:     shutdown,
		flush:        flush,
		Logger:       log.NewPrefixLogger(),
	}
	v.AddLogPrefix("view")
	v.AddLogPrefix("term")
	v.AddLogPrefix("http")
	go v.Run()
	return v
}

func (v *HttpView) Run() {
	updates := v.httpProto.Txns.Reg()

	for {
		select {
		case <-v.shutdown:
			return

		case txn := <-updates:
			v.Debug("Got HTTP update")
			if txn.(*proto.HttpTxn).Resp == nil {
				v.HttpRequests.Add(txn)
			}
			v.Render()
		}
	}
}

func (v *HttpView) Render() {
	v.Clear()
	v.Printf(0, 0, "HTTP Requests")
	v.Printf(0, 1, "-------------")
	for i, obj := range v.HttpRequests.Slice() {
		txn := obj.(*proto.HttpTxn)
		v.Printf(0, 3+i, "%s %v", txn.Req.Method, txn.Req.URL.Path)
		if txn.Resp != nil {
			v.APrintf(colorFor(txn.Resp.Status), 30, 3+i, "%s", txn.Resp.Status)
		}
	}
	v.flush <- 1
}

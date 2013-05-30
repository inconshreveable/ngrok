// interative web user interface
package web

import (
	"encoding/json"
	"encoding/xml"
	"html/template"
	"net/http"
	"net/http/httputil"
	"net/url"
	"ngrok/client/ui"
	"ngrok/client/views/web/static"
	"ngrok/log"
	"ngrok/proto"
	"ngrok/util"
	"strings"
)

type SerializedTxn struct {
	Id             string
	*proto.HttpTxn `json:"-"`
	Req            SerializedRequest
	Resp           SerializedResponse
}

type SerializedBody struct {
	RawContentType string
	ContentType    string
	Text           string
	Length         int
	Error          string
	ErrorOffset    int
        Form   url.Values
}

type SerializedRequest struct {
	Raw        string
	MethodPath string
	Params     url.Values
	Header     http.Header
	Body       SerializedBody
}

type SerializedResponse struct {
	Raw    string
	Status string
	Header http.Header
	Body   SerializedBody
}

type WebHttpView struct {
	webview      *WebView
	ctl          *ui.Controller
	httpProto    *proto.Http
	HttpRequests *util.Ring
	idToTxn      map[string]*SerializedTxn
}

func NewWebHttpView(wv *WebView, ctl *ui.Controller, proto *proto.Http) *WebHttpView {
	w := &WebHttpView{
		webview:      wv,
		ctl:          ctl,
		httpProto:    proto,
		idToTxn:      make(map[string]*SerializedTxn),
		HttpRequests: util.NewRing(20),
	}
	go w.update()
	w.register()
	return w
}

type XMLDoc struct {
	data []byte `xml:",innerxml"`
}

func makeBody(h http.Header, body []byte) SerializedBody {
	b := SerializedBody{
		Length:      len(body),
		Text:        string(body),
		ErrorOffset: -1,
	}

	// some errors like XML errors only give a line number
	// and not an exact offset
	offsetForLine := func(line int) int {
		lines := strings.SplitAfterN(b.Text, "\n", line)
		return b.Length - len(lines[len(lines)-1])
	}

	var err error
	b.RawContentType = h.Get("Content-Type")
	if b.RawContentType != "" {
		b.ContentType = strings.TrimSpace(strings.Split(b.RawContentType, ";")[0])
		switch b.ContentType {
		case "application/xml":
		case "text/xml":
			err = xml.Unmarshal(body, new(XMLDoc))
			if err != nil {
				syntaxError := err.(*xml.SyntaxError)
				// xml syntax errors only give us a line number, so we
				// count to find an offset
				b.ErrorOffset = offsetForLine(syntaxError.Line)
			}

		case "application/json":
			err = json.Unmarshal(body, new(json.RawMessage))
			if err != nil {
				syntaxError := err.(*json.SyntaxError)
				b.ErrorOffset = int(syntaxError.Offset)
			}

		case "application/x-www-form-urlencoded":
			b.Form, err = url.ParseQuery(string(body))
		}
	}

	if err != nil {
		b.Error = err.Error()
	}

	return b
}

func (whv *WebHttpView) update() {
	// open channels for incoming http state changes
	// and broadbasts
	txnUpdates := whv.httpProto.Txns.Reg()
	for {
		select {
		case txn := <-txnUpdates:
			// XXX: it's not safe for proto.Http and this code
			// to be accessing txn and txn.(req/resp) without synchronization
			htxn := txn.(*proto.HttpTxn)

                        // we haven't processed this transaction yet if we haven't set the
                        // user data
			if htxn.UserData == nil {
				id, err := util.RandId(8)
				if err != nil {
					log.Error("Failed to generate txn identifier for web storage: %v", err)
					continue
				}

				rawReq, err := httputil.DumpRequestOut(htxn.Req.Request, true)
				if err != nil {
					log.Error("Failed to dump request: %v", err)
					continue
				}

				body := makeBody(htxn.Req.Header, htxn.Req.BodyBytes)
				whtxn := &SerializedTxn{
					Id:      id,
					HttpTxn: htxn,
					Req: SerializedRequest{
						MethodPath: htxn.Req.Method + " " + htxn.Req.URL.Path,
						Raw:        string(rawReq),
						Params:     htxn.Req.URL.Query(),
						Header:     htxn.Req.Header,
						Body:       body,
					},
				}

				htxn.UserData = whtxn
				// XXX: unsafe map access from multiple go routines
				whv.idToTxn[whtxn.Id] = whtxn
				// XXX: use return value to delete from map so we don't leak memory
				whv.HttpRequests.Add(whtxn)
                        } else {
				rawResp, err := httputil.DumpResponse(htxn.Resp.Response, true)
				if err != nil {
					log.Error("Failed to dump response: %v", err)
					continue
				}

				txn := htxn.UserData.(*SerializedTxn)
				body := makeBody(htxn.Resp.Header, htxn.Resp.BodyBytes)
				txn.Resp = SerializedResponse{
					Status: htxn.Resp.Status,
					Raw:    string(rawResp),
					Header: htxn.Resp.Header,
					Body:   body,
				}

				payload, err := json.Marshal(txn)
				if err != nil {
					log.Error("Failed to serialized txn payload for websocket: %v", err)
				}
				whv.webview.wsMessages.In() <- payload
			}
		}
	}
}

func (h *WebHttpView) register() {
	http.HandleFunc("/http/in/replay", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		txnid := r.Form.Get("txnid")
		if txn, ok := h.idToTxn[txnid]; ok {
			bodyBytes, err := httputil.DumpRequestOut(txn.HttpTxn.Req.Request, true)
			if err != nil {
				panic(err)
			}
			h.ctl.Cmds <- ui.CmdRequest{Payload: bodyBytes}
			w.Write([]byte(http.StatusText(200)))
		} else {
			// XXX: 400
			http.NotFound(w, r)
		}
	})

	http.HandleFunc("/http/in", func(w http.ResponseWriter, r *http.Request) {
		/*
			funcMap := template.FuncMap{
				"handleForm": func(b []byte, h http.Header) (values interface{}, err error) {

					if b != nil {
						values, err = url.ParseQuery(string(b))
					}
					return
				},
			}
		*/

		tmpl := template.Must(template.New("page.html").Delims("{%", "%}").Parse(string(static.PageHtml())))

		payload, err := json.Marshal(h.HttpRequests.Slice())
		if err != nil {
			panic(err)
		}

		// write the response
		if err := tmpl.Execute(w, string(payload)); err != nil {
			panic(err)
		}
	})
}

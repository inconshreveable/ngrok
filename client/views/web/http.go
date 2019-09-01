// interactive web user interface
package web

import (
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"html/template"
	"net/http"
	"net/http/httputil"
	"net/url"
	"ngrok/client/assets"
	"ngrok/client/mvc"
	"ngrok/log"
	"ngrok/proto"
	"ngrok/util"
	"strings"
	"unicode/utf8"
)

type SerializedTxn struct {
	Id             string
	Duration       int64
	Start          int64
	ConnCtx        mvc.ConnectionContext
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
	Form           url.Values
}

type SerializedRequest struct {
	Raw        string
	MethodPath string
	Params     url.Values
	Header     http.Header
	Body       SerializedBody
	Binary     bool
}

type SerializedResponse struct {
	Raw    string
	Status string
	Header http.Header
	Body   SerializedBody
	Binary bool
}

type WebHttpView struct {
	log.Logger

	webview      *WebView
	ctl          mvc.Controller
	httpProto    *proto.Http
	state        chan SerializedUiState
	HttpRequests *util.Ring
	idToTxn      map[string]*SerializedTxn
}

type SerializedUiState struct {
	Tunnels []mvc.Tunnel
}

type SerializedPayload struct {
	Txns    []interface{}
	UiState SerializedUiState
}

func newWebHttpView(ctl mvc.Controller, wv *WebView, proto *proto.Http) *WebHttpView {
	whv := &WebHttpView{
		Logger:       log.NewPrefixLogger("view", "web", "http"),
		webview:      wv,
		ctl:          ctl,
		httpProto:    proto,
		idToTxn:      make(map[string]*SerializedTxn),
		HttpRequests: util.NewRing(20),
	}
	ctl.Go(whv.updateHttp)
	whv.register()
	return whv
}

type XMLDoc struct {
	data []byte `xml:",innerxml"`
}

func makeBody(h http.Header, body []byte) SerializedBody {
	b := SerializedBody{
		Length:      len(body),
		Text:        base64.StdEncoding.EncodeToString(body),
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
		case "application/xml", "text/xml":
			err = xml.Unmarshal(body, new(XMLDoc))
			if err != nil {
				if syntaxError, ok := err.(*xml.SyntaxError); ok {
					// xml syntax errors only give us a line number, so we
					// count to find an offset
					b.ErrorOffset = offsetForLine(syntaxError.Line)
				}
			}

		case "application/json":
			err = json.Unmarshal(body, new(json.RawMessage))
			if err != nil {
				if syntaxError, ok := err.(*json.SyntaxError); ok {
					b.ErrorOffset = int(syntaxError.Offset)
				}
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

func (whv *WebHttpView) updateHttp() {
	// open channels for incoming http state changes
	// and broadcasts
	txnUpdates := whv.httpProto.Txns.Reg()
	for txn := range txnUpdates {
		// XXX: it's not safe for proto.Http and this code
		// to be accessing txn and txn.(req/resp) without synchronization
		htxn := txn.(*proto.HttpTxn)

		// we haven't processed this transaction yet if we haven't set the
		// user data
		if htxn.UserCtx == nil {
			rawReq, err := proto.DumpRequestOut(htxn.Req.Request, true)
			if err != nil {
				whv.Error("Failed to dump request: %v", err)
				continue
			}

			body := makeBody(htxn.Req.Header, htxn.Req.BodyBytes)
			whtxn := &SerializedTxn{
				Id:      util.RandId(8),
				HttpTxn: htxn,
				Req: SerializedRequest{
					MethodPath: htxn.Req.Method + " " + htxn.Req.URL.Path,
					Raw:        base64.StdEncoding.EncodeToString(rawReq),
					Params:     htxn.Req.URL.Query(),
					Header:     htxn.Req.Header,
					Body:       body,
					Binary:     !utf8.Valid(rawReq),
				},
				Start:   htxn.Start.Unix(),
				ConnCtx: htxn.ConnUserCtx.(mvc.ConnectionContext),
			}

			htxn.UserCtx = whtxn
			// XXX: unsafe map access from multiple go routines
			whv.idToTxn[whtxn.Id] = whtxn
			// XXX: use return value to delete from map so we don't leak memory
			whv.HttpRequests.Add(whtxn)
		} else {
			rawResp, err := httputil.DumpResponse(htxn.Resp.Response, true)
			if err != nil {
				whv.Error("Failed to dump response: %v", err)
				continue
			}

			txn := htxn.UserCtx.(*SerializedTxn)
			body := makeBody(htxn.Resp.Header, htxn.Resp.BodyBytes)
			txn.Duration = htxn.Duration.Nanoseconds()
			txn.Resp = SerializedResponse{
				Status: htxn.Resp.Status,
				Raw:    base64.StdEncoding.EncodeToString(rawResp),
				Header: htxn.Resp.Header,
				Body:   body,
				Binary: !utf8.Valid(rawResp),
			}

			payload, err := json.Marshal(txn)
			if err != nil {
				whv.Error("Failed to serialized txn payload for websocket: %v", err)
			}
			whv.webview.wsMessages.In() <- payload
		}
	}
}

func (whv *WebHttpView) register() {
	http.HandleFunc("/http/in/replay", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				err := util.MakePanicTrace(r)
				whv.Error("Replay failed: %v", err)
				http.Error(w, err, 500)
			}
		}()

		r.ParseForm()
		txnid := r.Form.Get("txnid")
		if txn, ok := whv.idToTxn[txnid]; ok {
			reqBytes, err := base64.StdEncoding.DecodeString(txn.Req.Raw)
			if err != nil {
				panic(err)
			}
			whv.ctl.PlayRequest(txn.ConnCtx.Tunnel, reqBytes)
			w.Write([]byte(http.StatusText(200)))
		} else {
			http.Error(w, http.StatusText(400), 400)
		}
	})

	http.HandleFunc("/http/in", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				err := util.MakePanicTrace(r)
				whv.Error("HTTP web view failed: %v", err)
				http.Error(w, err, 500)
			}
		}()

		pageTmpl, err := assets.Asset("assets/client/page.html")
		if err != nil {
			panic(err)
		}

		tmpl := template.Must(template.New("page.html").Delims("{%", "%}").Parse(string(pageTmpl)))

		payloadData := SerializedPayload{
			Txns:    whv.HttpRequests.Slice(),
			UiState: SerializedUiState{Tunnels: whv.ctl.State().GetTunnels()},
		}

		payload, err := json.Marshal(payloadData)
		if err != nil {
			panic(err)
		}

		// write the response
		if err := tmpl.Execute(w, string(payload)); err != nil {
			panic(err)
		}
	})
}

func (whv *WebHttpView) Shutdown() {
}

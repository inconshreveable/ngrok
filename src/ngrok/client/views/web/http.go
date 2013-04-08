// interative web user interface
package web

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"ngrok/client/ui"
	"ngrok/client/views/web/static"
	"ngrok/proto"
	"ngrok/util"
	"strings"
)

func isContentType(h http.Header, ctypes ...string) bool {
	for _, ctype := range ctypes {
		if strings.Contains(h.Get("Content-Type"), ctype) {
			return true
		}
	}
	return false
}

func readBody(r io.Reader) ([]byte, io.ReadCloser, error) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r)
	return buf.Bytes(), ioutil.NopCloser(buf), err
}

type WebHttpTxn struct {
	Id string
	*proto.HttpTxn
	reqBody  []byte
	respBody []byte
}

type WebHttpView struct {
	ctl          *ui.Controller
	httpProto    *proto.Http
	HttpRequests *util.Ring
	idToTxn      map[string]*WebHttpTxn
}

func NewWebHttpView(ctl *ui.Controller, proto *proto.Http) *WebHttpView {
	w := &WebHttpView{
		ctl:          ctl,
		httpProto:    proto,
		idToTxn:      make(map[string]*WebHttpTxn),
		HttpRequests: util.NewRing(20),
	}
	go w.update()
	w.register()
	return w
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

			// XXX: golang, why do I have to do this to make DumpRequestOut work later?
			htxn.Req.URL.Scheme = "http"

			if htxn.Resp == nil {
				whtxn := &WebHttpTxn{Id: util.RandId(), HttpTxn: htxn}

				// XXX: unsafe map access from multiple go routines
				whv.idToTxn[whtxn.Id] = whtxn
				// XXX: use return value to delete from map so we don't leak memory
				whv.HttpRequests.Add(whtxn)
			}
		}
	}
}

func (h *WebHttpView) register() {
	http.HandleFunc("/http/in/replay", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		txnid := r.Form.Get("txnid")
		if txn, ok := h.idToTxn[txnid]; ok {
			bodyBytes, err := httputil.DumpRequestOut(txn.Req.Request, true)
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
		funcMap := template.FuncMap{
			"classForStatus": func(status string) string {
				switch status[0] {
				case '2':
					return "text-info"
				case '3':
					return "muted"
				case '4':
					return "text-warning"
				case '5':
					return "text-error"
				}
				return ""
			},
			"dumpResponse": func(resp *proto.HttpResponse) (interface{}, error) {
				b, err := httputil.DumpResponse(resp.Response, true)
				return string(b), err
			},
			"dumpRequest": func(req *proto.HttpRequest) (interface{}, error) {
				b, err := httputil.DumpRequestOut(req.Request, true)
				return string(b), err
			},
			"handleForm": func(b []byte, h http.Header) (values interface{}, err error) {
				if !isContentType(h, "application/x-www-form-urlencoded") {
					return
				}

				if b != nil {
					values, err = url.ParseQuery(string(b))
				}
				return
			},
			"handleJson": func(raw []byte, h http.Header) interface{} {
				if !isContentType(h, "application/json") {
					return nil
				}

				var err error
				pretty := new(bytes.Buffer)
				out := raw
				if raw != nil {
					err = json.Indent(pretty, raw, "", "    ")
					if err == nil {
						out = pretty.Bytes()
					}
				}

				return struct {
					Str string
					Err error
				}{
					string(out),
					err,
				}
			},
			"handleOther": func(b []byte, h http.Header) interface{} {
				if isContentType(h, "application/json", "application/x-www-form-urlencoded") {
					return nil
				}

				return string(b)
			},
		}

		tmpl := template.Must(
			template.New("page.html").Funcs(funcMap).Parse(string(static.PageHtml())))
		template.Must(tmpl.Parse(string(static.BodyHtml())))

		// write the response
		if err := tmpl.Execute(w, h); err != nil {
			panic(err)
		}
	})
}

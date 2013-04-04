// interative web user interface
package web

import (
	"bytes"
	"encoding/json"
	"html/template"
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

func readBody(r *http.Request) ([]byte, error) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	r.Body = ioutil.NopCloser(buf)
	return buf.Bytes(), err
}

type WebHttpTxn struct {
	Id string
	*proto.HttpTxn
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
			bodyBytes, err := httputil.DumpRequestOut(txn.Req, true)
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
			"dumpResponse": func(resp *http.Response) (interface{}, error) {
				b, err := httputil.DumpResponse(resp, true)
				return string(b), err
			},
			"dumpRequest": func(req *http.Request) (interface{}, error) {
				b, err := httputil.DumpRequestOut(req, true)
				return string(b), err
			},
			"handleForm": func(req *http.Request) (values interface{}, err error) {
				if !strings.Contains(req.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
					return
				}

				b, err := readBody(req)
				if err != nil {
					return
				}

				values, err = url.ParseQuery(string(b))
				return
			},
			"handleJson": func(req *http.Request) interface{} {
				if !strings.Contains(req.Header.Get("Content-Type"), "application/json") {
					return nil
				}

				raw, err := readBody(req)
				if err != nil {
					panic(err)
				}

				pretty := new(bytes.Buffer)
				err = json.Indent(pretty, raw, "", "    ")

				retval := struct {
					Str string
					Err error
				}{
					string(pretty.Bytes()),
					err,
				}

				if err != nil {
					retval.Str = string(raw)
				}

				return retval
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

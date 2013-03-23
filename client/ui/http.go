// interative http client user interface
package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type RepeatableReader struct {
	io.Reader
	buffer []byte
}

func NewRepeatableReader(rd io.ReadCloser) *RepeatableReader {
	buffer := new(bytes.Buffer)
	buffer.ReadFrom(rd)
	return &RepeatableReader{
		bytes.NewBuffer(buffer.Bytes()),
		buffer.Bytes(),
	}
}

func (rr *RepeatableReader) Read(b []byte) (n int, err error) {
	n, err = rr.Reader.Read(b)

	if err == io.EOF {
		rr.Reader = bytes.NewBuffer(rr.buffer)
	}

	return n, err
}

func (rr *RepeatableReader) Close() error {
	return nil
}

type Http struct {
	ui   *Ui
	port int
}

func NewHttp(port int) *Http {
	return &Http{port: port}
}

func (h *Http) SetUi(ui *Ui) {
	h.ui = ui
	go h.run()
}

func (h *Http) run() {
	// open channels for incoming application state changes
	// and broadbasts

	updates := h.ui.Updates.Reg()
	var s State
	go func() {
		for {
			select {
			case obj := <-updates:
				s = obj.(State)
			}
		}
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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
				b, err := httputil.DumpResponse(resp, false)
				body := new(bytes.Buffer)
				body.ReadFrom(resp.Body)
				return string(b) + string(body.Bytes()), err
			},
			"dumpRequest": func(req *http.Request) (interface{}, error) {
				b, err := httputil.DumpRequest(req, false)
				body := new(bytes.Buffer)
				body.ReadFrom(req.Body)
				return string(b) + string(body.Bytes()), err
			},
			"handleForm": func(req *http.Request) (values interface{}, err error) {
				if req.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
					return
				}

				b, err := ioutil.ReadAll(req.Body)
				if err != nil {
					return
				}

				values, err = url.ParseQuery(string(b))
				return
			},
			"handleJson": func(req *http.Request) interface{} {
				if req.Header.Get("Content-Type") != "application/json" {
					return nil
				}

				src := new(bytes.Buffer)
				dst := new(bytes.Buffer)
				src.ReadFrom(req.Body)
				err := json.Indent(dst, src.Bytes(), "", "    ")

				retval := struct {
					Str string
					Err error
				}{
					string(dst.Bytes()),
					err,
				}

				if err != nil {
					retval.Str = string(src.Bytes())
				}

				return retval
			},
		}

		tmpl := template.Must(
			template.New("page.html").Funcs(funcMap).ParseFiles("./templates/page.html", "./templates/body.html"))

		for _, htxn := range s.GetHistory() {
			req, resp := htxn.GetRequest(), htxn.GetResponse()

			req.Body = NewRepeatableReader(req.Body)
			if resp != nil && resp.Body != nil {
				resp.Body = NewRepeatableReader(resp.Body)
			}
		}

		// write the response
		if err := tmpl.Execute(w, s); err != nil {
			panic(err)
		}
	})

	http.ListenAndServe(fmt.Sprintf(":%d", h.port), nil)
}

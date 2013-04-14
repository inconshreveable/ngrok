// interative web user interface
package web

import (
	"fmt"
	"net/http"
	"ngrok/client/ui"
	"ngrok/client/views/web/static"
	"ngrok/log"
	"ngrok/proto"
	"strings"
)

type WebView struct{}

func NewWebView(ctl *ui.Controller, state ui.State, port int) *WebView {
	w := &WebView{}

	switch p := state.GetProtocol().(type) {
	case *proto.Http:
		NewWebHttpView(ctl, p)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/http/in", 302)
	})

	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		name := parts[len(parts)-1]
		fn, ok := static.AssetMap[name]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Write(fn())
	})

	log.Info("Serving web interface on localhost:%d", port)
	go http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	return w
}

// interative web user interface
package ui

import (
	"fmt"
	"net/http"
	"ngrok/proto"
)

type WebView struct{}

func NewWebView(ctl *Controller, state State, port int) *WebView {
	w := &WebView{}

	switch p := state.GetProtocol().(type) {
	case *proto.Http:
		NewWebHttpView(ctl, p)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/http/in", 302)
	})

	go http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	return w
}

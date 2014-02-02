// interactive web user interface
package web

import (
	"github.com/gorilla/websocket"
	"net/http"
	"ngrok/client/assets"
	"ngrok/client/mvc"
	"ngrok/log"
	"ngrok/proto"
	"ngrok/util"
	"path"
)

type WebView struct {
	log.Logger

	ctl mvc.Controller

	// messages sent over this broadcast are sent to all websocket connections
	wsMessages *util.Broadcast
}

func NewWebView(ctl mvc.Controller, addr string) *WebView {
	wv := &WebView{
		Logger:     log.NewPrefixLogger("view", "web"),
		wsMessages: util.NewBroadcast(),
		ctl:        ctl,
	}

	// for now, always redirect to the http view
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/http/in", 302)
	})

	// handle web socket connections
	http.HandleFunc("/_ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Upgrade(w, r, nil, 1024, 1024)

		if err != nil {
			http.Error(w, "Failed websocket upgrade", 400)
			wv.Warn("Failed websocket upgrade: %v", err)
			return
		}

		msgs := wv.wsMessages.Reg()
		defer wv.wsMessages.UnReg(msgs)
		for m := range msgs {
			err := conn.WriteMessage(websocket.TextMessage, m.([]byte))
			if err != nil {
				// connection is closed
				break
			}
		}
	})

	// serve static assets
	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		buf, err := assets.Asset(path.Join("assets", "client", r.URL.Path[1:]))
		if err != nil {
			wv.Warn("Error serving static file: %s", err.Error())
			http.NotFound(w, r)
			return
		}
		w.Write(buf)
	})

	wv.Info("Serving web interface on %s", addr)
	wv.ctl.Go(func() { http.ListenAndServe(addr, nil) })
	return wv
}

func (wv *WebView) NewHttpView(proto *proto.Http) *WebHttpView {
	return newWebHttpView(wv.ctl, wv, proto)
}

func (wv *WebView) Shutdown() {
}

/*
   interactive terminal interface for local clients
*/
package term

import (
	"fmt"
	termbox "github.com/nsf/termbox-go"
	"ngrok/client/ui"
	"ngrok/log"
	"ngrok/proto"
	"time"
)

type TermView struct {
	ctl      *ui.Controller
	updates  chan interface{}
	flush    chan int
	subviews []ui.View
	state    ui.State
	log.Logger
	*area
}

func New(ctl *ui.Controller, state ui.State) *TermView {
	// initialize terminal display
	termbox.Init()

	// make sure ngrok doesn't quit until we've cleaned up
	ctl.Wait.Add(1)

	w, _ := termbox.Size()

	v := &TermView{
		ctl:      ctl,
		updates:  ctl.Updates.Reg(),
		flush:    make(chan int),
		subviews: make([]ui.View, 0),
		state:    state,
		Logger:   log.NewPrefixLogger(),
		area:     NewArea(0, 0, w, 10),
	}

	v.Logger.AddLogPrefix("view")
	v.Logger.AddLogPrefix("term")

	switch p := state.GetProtocol().(type) {
	case *proto.Http:
		v.subviews = append(v.subviews, NewHttp(p, v.flush, ctl.Shutdown, 0, 10))
	default:
	}

	v.Render()

	go v.run()
	go v.input()

	return v
}

func colorForConn(status string) termbox.Attribute {
	switch status {
	case "connecting":
		return termbox.ColorCyan
	case "reconnecting":
		return termbox.ColorRed
	case "online":
		return termbox.ColorGreen
	}
	return termbox.ColorWhite
}

func (v *TermView) Render() {
	v.Clear()

	// quit instructions
	quitMsg := "(Ctrl+C to quit)"
	v.Printf(v.w-len(quitMsg), 0, quitMsg)

	// new version message
	newVersion := v.state.GetNewVersion()
	if newVersion != "" {
		newVersionMsg := fmt.Sprintf("new version available at http://ngrok.com")
		v.APrintf(termbox.ColorYellow, 30, 0, newVersionMsg)
	}

	v.APrintf(termbox.ColorBlue|termbox.AttrBold, 0, 0, "ngrok")

	status := v.state.GetStatus()
	v.APrintf(colorForConn(status), 0, 2, "%-30s%s", "Tunnel Status", status)

	v.Printf(0, 3, "%-30s%s/%s", "Version", v.state.GetClientVersion(), v.state.GetServerVersion())
	v.Printf(0, 4, "%-30s%s", "Protocol", v.state.GetProtocol().GetName())
	v.Printf(0, 5, "%-30s%s -> %s", "Forwarding", v.state.GetPublicUrl(), v.state.GetLocalAddr())
	webAddr := fmt.Sprintf("http://localhost:%d", v.state.GetWebPort())
	v.Printf(0, 6, "%-30s%s", "Web Interface", webAddr)

	connMeter, connTimer := v.state.GetConnectionMetrics()
	v.Printf(0, 7, "%-30s%d", "# Conn", connMeter.Count())

	msec := float64(time.Millisecond)
	v.Printf(0, 8, "%-30s%.2fms", "Avg Conn Time", connTimer.Mean()/msec)

	termbox.Flush()
}

func (v *TermView) run() {
	defer v.ctl.Wait.Done()
	defer termbox.Close()

	for {
		v.Debug("Waiting for update")
		select {
		case <-v.flush:
			termbox.Flush()

		case obj := <-v.updates:
			if obj != nil {
				v.state = obj.(ui.State)
			}
			v.Render()

		case <-v.ctl.Shutdown:
			return
		}
	}
}

func (v *TermView) input() {
	for {
		ev := termbox.PollEvent()
		switch ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyCtrlC:
				v.Info("Got quit command")
				v.ctl.Cmds <- ui.CmdQuit{}
			}

		case termbox.EventResize:
			v.Info("Resize event, redrawing")
			// send nil to update channel to force re-rendering
			v.updates <- nil
			for _, sv := range v.subviews {
				sv.Render()
			}

		case termbox.EventError:
			if v.ctl.IsShuttingDown() {
				return
			}
			panic(ev.Err)
		}
	}
}

/* 
   interactive terminal interface for local clients
*/
package ui

import (
	"fmt"
	termbox "github.com/nsf/termbox-go"
	"time"
)

const (
	fgColor = termbox.ColorWhite
	bgColor = termbox.ColorDefault
)

func clear() {
	w, h := termbox.Size()

	for i := 0; i < w; i++ {
		for j := 0; j < h; j++ {
			termbox.SetCell(i, j, ' ', fgColor, bgColor)
		}
	}
}

func printfAttr(x, y int, fg termbox.Attribute, arg0 string, args ...interface{}) {
	s := fmt.Sprintf(arg0, args...)
	for i, ch := range s {
		termbox.SetCell(x+i, y, ch, fg, bgColor)
	}
}

func printf(x, y int, arg0 string, args ...interface{}) {
	printfAttr(x, y, fgColor, arg0, args...)
}

type TermView struct {
	ctl            *Controller
	statusColorMap map[string]termbox.Attribute
	updates        chan interface{}
}

func NewTermView(ctl *Controller) *TermView {
	t := &TermView{
		ctl:     ctl,
		updates: ctl.Updates.Reg(),
		statusColorMap: map[string]termbox.Attribute{
			"connecting":   termbox.ColorCyan,
			"reconnecting": termbox.ColorRed,
			"online":       termbox.ColorGreen,
		},
	}

	go t.run()
	return t
}

func (t *TermView) run() {
	// XXX: clean this up? maybe Term should have its own waitgroup for
	// both run and draw

	// make sure we shut down cleanly
	t.ctl.Wait.Add(1)
	defer t.ctl.Wait.Done()

	// init/close termbox library
	termbox.Init()
	defer termbox.Close()

	go t.input()

	t.draw()
}

func (t *TermView) draw() {
	var state State
	for {
		select {
		case newState := <-t.updates:
			if newState != nil {
				state = newState.(State)
			}

			if state == nil {
				// log.Info("Got update to draw, but no state to draw with")
				continue
			}

			// program is shutting down
			if state.IsStopping() {
				return
			}

			clear()

			x, _ := termbox.Size()
			quitMsg := "(Ctrl+C to quit)"
			printf(x-len(quitMsg), 0, quitMsg)

			printfAttr(0, 0, termbox.ColorBlue|termbox.AttrBold, "ngrok")

			msec := float64(time.Millisecond)

			printfAttr(0, 2, t.statusColorMap[state.GetStatus()], "%-30s%s", "Tunnel Status", state.GetStatus())
			printf(0, 3, "%-30s%s", "Version", state.GetVersion())
			printf(0, 4, "%-30s%s", "Protocol", state.GetProtocol().GetName())
			printf(0, 5, "%-30s%s -> %s", "Forwarding", state.GetPublicUrl(), state.GetLocalAddr())
			printf(0, 6, "%-30s%s", "HTTP Dashboard", "http://127.0.0.1:9999")

			connMeter, connTimer := state.GetConnectionMetrics()
			printf(0, 7, "%-30s%d", "# Conn", connMeter.Count())
			printf(0, 8, "%-30s%.2fms", "Avg Conn Time", connTimer.Mean()/msec)

			/*
				if state.GetProtocol() == "http" {
					printf(0, 10, "HTTP Requests")
					printf(0, 11, "-------------")
					for i, http := range state.GetHistory() {
						req := http.GetRequest()
						resp := http.GetResponse()
						printf(0, 12+i, "%s %v", req.Method, req.URL)
						if resp != nil {
							printf(30, 12+i, "%s", resp.Status)
						}
					}
				}
			*/

			termbox.Flush()
		}
	}
}

func (t *TermView) input() {
	for {
		ev := termbox.PollEvent()
		switch ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyCtrlC:
				t.ctl.Cmds <- Command{QUIT, ""}
				return
			}

		case termbox.EventResize:
			t.updates <- nil

		case termbox.EventError:
			panic(ev.Err)
		}
	}
}

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

type Term struct {
	ui             *Ui
	statusColorMap map[string]termbox.Attribute
}

func NewTerm() *Term {
	return &Term{
		statusColorMap: map[string]termbox.Attribute{
			"connecting":   termbox.ColorCyan,
			"reconnecting": termbox.ColorRed,
			"online":       termbox.ColorGreen,
		},
	}
}

func (t *Term) SetUi(ui *Ui) {
	t.ui = ui
	go t.run()
}

func (t *Term) run() {
	// make sure we shut down cleanly
	t.ui.Wait.Add(1)
	defer t.ui.Wait.Done()

	// open channels for incoming application state changes
	// and broadbasts
	updates := t.ui.Updates.Reg()

	// init/close termbox library
	termbox.Init()
	defer termbox.Close()

	go t.input()

	t.draw(updates)
}

func (t *Term) draw(updates chan State) {
	for {
		select {
		case state := <-updates:
			// program is shutting down
			if state.IsStopping() {
				return
			}

			clear()

			printfAttr(0, 0, termbox.ColorBlue|termbox.AttrBold, "ngrok")

			msec := float64(time.Millisecond)

			printf(0, 2, "%-30s%s", "Version", state.GetVersion())
			printf(0, 3, "%-30s%s", "Public URL", state.GetPublicUrl())
			printf(0, 4, "%-30s%s", "Local Address", state.GetLocalAddr())
			printfAttr(0, 5, t.statusColorMap[state.GetStatus()], "%-30s%s", "Tunnel Status", state.GetStatus())

			connMeter, connTimer := state.GetConnectionMetrics()
			printf(0, 6, "%-30s%d", "# Conn", connMeter.Count())
			printf(0, 7, "%-30s%.2fms", "Mean Conn Time", connTimer.Mean()/msec)
			printf(0, 8, "%-30s%.2fms", "Conn Time 95th PCTL", connTimer.Percentile(0.95)/msec)

			bytesInCount, bytesIn := state.GetBytesInMetrics()
			printf(0, 9, "%-30s%d", "Bytes In", bytesInCount.Count())
			printf(0, 10, "%-30s%.2f", "Avg Bytes/req", bytesIn.Mean())

			bytesOutCount, bytesOut := state.GetBytesOutMetrics()
			printf(0, 11, "%-30s%d", "Bytes Out", bytesOutCount.Count())
			printf(0, 12, "%-30s%.2f", "Bytes Out/req", bytesOut.Mean())

			printf(0, 14, "Last HTTP Requests")
			for i, http := range state.GetHistory() {
				req := http.GetRequest()
				resp := http.GetResponse()
				printf(0, 15+i, "%s %v", req.Method, req.URL)
				if resp != nil {
					printf(30, 15+i, "%s", resp.Status)
				}
			}

			termbox.Flush()
		}
	}
}

func (t *Term) input() {
	for {
		ev := termbox.PollEvent()
		switch ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyCtrlC:
				t.ui.Cmds <- QUIT
				return
			}
		}
	}
}

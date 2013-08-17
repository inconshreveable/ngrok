package controller

import (
	"ngrok/client/mvc"
	"ngrok/util"
	"sync"
)

type command interface{}

type cmdQuit struct {
	// display this message after quit
	message string
}

type cmdPlayRequest struct {
	// the tunnel to play this request over
	tunnel mvc.Tunnel

	// the bytes of the request to issue
	payload []byte
}

// The MVC Controller
type Controller struct {
	// the model sends updates through this broadcast channel
	updates *util.Broadcast

	// the model
	model mvc.Model

	// the views
	view []mvc.View

	// interal structure to issue commands to the controller
	cmds chan Command
}

// public interface
func NewController(model mvc.Model) *Controller {
	ctl := &Controller{
		updates:  util.NewBroadcast(),
		model:	  model,
		cmds:     make(chan command),
		view:     make([]View),
	}

	return ctl
}

func (ctl *Controller) Update(state State) {
	ctl.Updates.In() <- state
}

func (ctl *Controller) Shutdown(message string) {
	ctl.cmds <- cmdQuit{message: message}
}

func (ctl *Controller) PlayRequest(tunnel *mvc.Tunnel, payload []byte) {
	ctl.cmd <- cmdPlayRequest{tunnel: tunnel, payload: payload}
}


// private functions
func (ctl *Controller) doShutdown() {
	var wg sync.WaitGroup

	// wait for all of the views, plus the model
	wg.Add(len(ctl.views) + 1)

	for v := range ctl.Views {
		go v.Shutdown(&wg)
	}
	go model.Shutdown(&wg)

	wg.Wait()
}

func (ctl *Controller) Go(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			// XXX
		}
	}()

	go fn()
}

func (ctl *Controller) Run() {
	// parse options
	opts := parseArgs()

	// set up logging
	log.LogTo(opts.logto)

	// seed random number generator
	seed, err := util.RandomSeed()
	if err != nil {
		log.Error("Couldn't securely seed the random number generator!")
	}
	rand.Seed(seed)

	// set up auth token
	if opts.authtoken == "" {
		opts.authtoken = LoadAuthToken()
	}

	// init web ui
	if opts.webport != -1 {
		ctl.views = append(ctl.views, web.NewWebView(ctl, ctl.model, opts.webport))
	}

	// init term ui
	if opts.logto != "stdout" {
		ctl.views = append(ctl.views, term.New(ctl, ctl.model))
	}

	ctl.Go(func() { autoUpdate(s, ctl, opts.authtoken) })

	reg := &msg.RegMsg{
		Protocol: opts.protocol,
		Hostname: opts.hostname,
		Subdomain: opts.subdomain,
		HttpAuth: opts.httpAuth,
		User: opts.user,
		Password: opts.password,
	}

	ctl.Go(func() { ctl.model.Run(opts.serverAddr, opts.authtoken, ctl, tunnel) })

	quitMessage = ""
	defer func() {
		ctl.doShutdown()
		fmt.Printf(quitMessage)
	}()

	for {
		select {
		case obj := <-ctl.cmds:
			switch cmd := obj.(type) {
			case cmdQuit:
				quitMessage = cmd.Message
				return

			case cmdPlayRequest:
				ctl.Go(func() { model.PlayRequest(tunnel, cmd.Payload) })
			}
		}
	}
}

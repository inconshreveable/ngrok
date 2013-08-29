package client

import (
	"math/rand"
	"ngrok/log"
	"ngrok/util"
)

func Main() {
	// parse options
	opts := parseArgs()

	// set up logging
	log.LogTo(opts.logto)

	// set up auth token
	if opts.authtoken == "" {
		opts.authtoken = LoadAuthToken()
	}

	// seed random number generator
	seed, err := util.RandomSeed()
	if err != nil {
		log.Error("Couldn't securely seed the random number generator!")
	}
	rand.Seed(seed)

	NewController().Run(opts)
}

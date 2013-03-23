.PHONY: default server client deps clean all

default: all

deps:
	GOPATH=`pwd` go get -v ngrok/...

server: deps
	GOPATH=`pwd` go install main/ngrokd

client: deps
	GOPATH=`pwd` go install main/ngrok

all: deps
	GOPATH=`pwd` go install main/...

clean:
	GOPATH=`pwd` go clean ngrok/...

.PHONY: default server client deps fmt clean all release-client release-server release-all
BUILDTAGS=
export GOPATH:=$(shell pwd)

default: all

deps:
	go get -tags '$(BUILDTAGS)' -d -v ngrok/...

server: deps
	go install -tags '$(BUILDTAGS)' ngrok/server/main

fmt:
	go fmt ngrok/...

client: deps
	go install -tags '$(BUILDTAGS)' ngrok/client/main

release-client: BUILDTAGS=release
release-client: client
	./nrsc ./bin/ngrok ./assets/client

release-server: BUILDTAGS=release
release-server: server
	./nrsc ./bin/ngrokd ./assets/server

release-all: release-client release-server

all: fmt client server

clean:
	go clean ngrok/...

.PHONY: default server client deps fmt clean all release-client release-server release-all client-assets server-assets
export GOPATH:=$(shell pwd)

default: all

deps:
	go get -tags '$(BUILDTAGS)' -d -v github.com/inconshreveable/ngrok/...

server: deps
	go install -tags '$(BUILDTAGS)' github.com/inconshreveable/ngrok/main/ngrokd

fmt:
	go fmt github.com/inconshreveable/ngrok/...

client: deps
	go install -tags '$(BUILDTAGS)' github.com/inconshreveable/ngrok/main/ngrok

client-assets:
	go get github.com/inconshreveable/go-bindata
	GOOS="" GOARCH="" go install github.com/inconshreveable/go-bindata
	bin/go-bindata -o src/github.com/inconshreveable/ngrok/client/assets assets/client

server-assets:
	go get github.com/inconshreveable/go-bindata
	GOOS="" GOARCH="" go install github.com/inconshreveable/go-bindata
	bin/go-bindata -o src/github.com/inconshreveable/ngrok/server/assets assets/server

release-client: BUILDTAGS=release
release-client: client-assets client

release-server: BUILDTAGS=release
release-server: server-assets server

release-all: release-client release-server

all: fmt client server

clean:
	go clean github.com/inconshreveable/ngrok/...

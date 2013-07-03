.PHONY: default server client deps fmt clean all release-client release-server release-all client-assets server-assets
BUILDTAGS=
export GOPATH:=$(shell pwd)

default: all

deps:
	go get -tags '$(BUILDTAGS)' -d -v ngrok/...

server: deps
	go install -tags '$(BUILDTAGS)' ngrok/main/ngrokd

fmt:
	go fmt ngrok/...

client: deps
	go install -tags '$(BUILDTAGS)' ngrok/main/ngrok

client-assets:
	go install github.com/inconshreveable/go-bindata
	bin/go-bindata -o src/ngrok/client/assets assets/client

server-assets:
	go install github.com/inconshreveable/go-bindata
	bin/go-bindata -o src/ngrok/server/assets assets/server

release-client: BUILDTAGS=release
release-client: client-assets client

release-server: BUILDTAGS=release
release-server: server-assets server

release-all: release-client release-server

all: fmt client server

clean:
	go clean ngrok/...

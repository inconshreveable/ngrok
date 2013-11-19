.PHONY: default server client deps fmt clean all release-client release-server release-all client-assets server-assets contributors
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
	go get github.com/inconshreveable/go-bindata
	GOOS="" GOARCH="" go install github.com/inconshreveable/go-bindata
	bin/go-bindata -o src/ngrok/client/assets assets/client

server-assets:
	go get github.com/inconshreveable/go-bindata
	GOOS="" GOARCH="" go install github.com/inconshreveable/go-bindata
	bin/go-bindata -o src/ngrok/server/assets assets/server

release-client: BUILDTAGS=release
release-client: client-assets client

release-server: BUILDTAGS=release
release-server: server-assets server

release-all: release-client release-server

all: fmt client server

clean:
	go clean -i -r ngrok/...

contributors:
	echo "Contributors to ngrok, both large and small:\n" > CONTRIBUTORS
	git log --raw | grep "^Author: " | sort | uniq | cut -d ' ' -f2- | sed 's/^/- /' | cut -d '<' -f1 >> CONTRIBUTORS

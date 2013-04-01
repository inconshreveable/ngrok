.PHONY: default server client deps clean all release-client release-server bindata
BUILDTAGS=
export GOPATH:=$(shell pwd)

default: all

deps:
	go get -v ngrok/...

server: deps
	go install -tags '$(BUILDTAGS)' main/ngrokd

client: deps
	go install -tags '$(BUILDTAGS)' main/ngrok

release-client: BUILDTAGS=release
release-client: bindata client

release-server: BUILDTAGS=release
release-server: server

release-all: release-client release-server

bindata:
	echo $$GOPATH
	go get github.com/inconshreveable/go-bindata
	./bin/go-bindata -b release -i templates/page.html -o src/ngrok/client/views/web/static/page.html.go -m -p static -f PageHtml
	./bin/go-bindata -b release -i templates/body.html -o src/ngrok/client/views/web/static/body.html.go -m -p static -f BodyHtml

all: client server

clean:
	go clean ngrok/...

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
	go get github.com/inconshreveable/go-bindata
	./bin/go-bindata -b release -i assets/page.html -o src/ngrok/client/views/web/static/page.html.go -m -p static -f PageHtml
	./bin/go-bindata -b release -i assets/body.html -o src/ngrok/client/views/web/static/body.html.go -m -p static -f BodyHtml
	./bin/go-bindata -b release -i assets/highlight.min.css -o src/ngrok/client/views/web/static/highlight.css.go -m -p static -f HighlightCss
	./bin/go-bindata -b release -i assets/highlight.min.js -o src/ngrok/client/views/web/static/highlight.js.go -m -p static -f HighlightJs
	./bin/go-bindata -b release -i assets/bootstrap.min.css -o src/ngrok/client/views/web/static/bootstrap.css.go -m -p static -f BootstrapCss
	./bin/go-bindata -b release -i assets/jquery-1.9.1.min.js -o src/ngrok/client/views/web/static/jquery-1.9.1.js.go -m -p static -f JqueryJs
	./bin/go-bindata -b release -i assets/vkbeautify.0.99.js -o src/ngrok/client/views/web/static/vkbeautify.0.99.js.go -m -p static -f VkBeautifyJs

all: client server

clean:
	go clean ngrok/...

.PHONY: default server client deps clean all release-client release-server bindata
BUILDTAGS=
export GOPATH:=$(shell pwd)

default: all

deps:
	go get -d -v ngrok/...

server: deps
	go install -tags '$(BUILDTAGS)' main/ngrokd

fmt:
	go fmt ngrok/...

client: deps
	go install -tags '$(BUILDTAGS)' main/ngrok

release-client: BUILDTAGS=release
release-client: bindata-client client

release-server: BUILDTAGS=release
release-server: server

release-all: release-client release-server

certs:
	go get github.com/inconshreveable/go-bindata
	./bin/go-bindata -i assets/tls/snakeoil.crt -o src/ngrok/server/tls/snakeoil.crt.go -m -p tls -f snakeoilCrt
	./bin/go-bindata -i assets/tls/snakeoil.key -o src/ngrok/server/tls/snakeoil.key.go -m -p tls -f snakeoilKey
	./bin/go-bindata -i assets/tls/snakeoilca.crt -o src/ngrok/client/tls/snakeoilca.crt.go -m -p tls -f snakeoilCaCrt
	./bin/go-bindata -i assets/tls/ngrokroot.crt -o src/ngrok/client/tls/ngrokroot.crt.go -m -p tls -f ngrokRootCrt

bindata-client:
	go get github.com/inconshreveable/go-bindata
	./bin/go-bindata -b release -i assets/page.html -o src/ngrok/client/views/web/static/page.html.go -m -p static -f PageHtml
	./bin/go-bindata -b release -i assets/highlight.min.css -o src/ngrok/client/views/web/static/highlight.css.go -m -p static -f HighlightCss
	./bin/go-bindata -b release -i assets/highlight.min.js -o src/ngrok/client/views/web/static/highlight.js.go -m -p static -f HighlightJs
	./bin/go-bindata -b release -i assets/bootstrap.min.css -o src/ngrok/client/views/web/static/bootstrap.css.go -m -p static -f BootstrapCss
	./bin/go-bindata -b release -i assets/jquery-1.9.1.min.js -o src/ngrok/client/views/web/static/jquery-1.9.1.js.go -m -p static -f JqueryJs
	./bin/go-bindata -b release -i assets/vkbeautify.js -o src/ngrok/client/views/web/static/vkbeautify.js.go -m -p static -f VkBeautifyJs
	./bin/go-bindata -b release -i assets/angular.js -o src/ngrok/client/views/web/static/angular.js.go -m -p static -f AngularJs
	./bin/go-bindata -b release -i assets/ngrok.js -o src/ngrok/client/views/web/static/ngrok.js.go -m -p static -f NgrokJs
	./bin/go-bindata -b release -i assets/base64.js -o src/ngrok/client/views/web/static/base64.js.go -m -p static -f Base64Js

all: fmt client server

clean:
	go clean ngrok/...

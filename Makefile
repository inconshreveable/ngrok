.PHONY: default server client deps fmt clean all release-all assets client-assets server-assets contributors

default: all

deps: assets
	go get -d -v ./...

compress:
	upx ngrokd
	upx ngrok

server: deps
	go build -o ngrokd ./cmd/ngrokd

fmt:
	go fmt ./...

client: deps
	go build -o ngrok ./cmd/ngrok

assets: client-assets server-assets

bin/go-bindata:
	go get -u github.com/jteeuwen/go-bindata/go-bindata

client-assets: bin/go-bindata
	go-bindata -nomemcopy -pkg=assets -tags=$(BUILDTAGS) \
		-o=client/assets/all.go \
		assets/client/...

server-assets: bin/go-bindata
	go-bindata -nomemcopy -pkg=assets -tags=$(BUILDTAGS) \
		-o=server/assets/all.go \
		assets/server/...

all: fmt client server

clean:
	go clean -i -r ./...
	rm -rf client/assets/ server/assets/

contributors:
	echo "Contributors to ngrok, both large and small:\n" > CONTRIBUTORS
	git log --raw | grep "^Author: " | sort | uniq | cut -d ' ' -f2- | sed 's/^/- /' | cut -d '<' -f1 >> CONTRIBUTORS

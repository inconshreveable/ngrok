.PHONY: default server client deps fmt clean all release-all assets client-assets server-assets check-authToken check-domain
export GOPATH:=$(shell pwd)

BUILDTAGS=debug
default: all

client-key:
	@echo "package client\n" > src/surf/client/key.go
	@echo "const (" >> src/surf/client/key.go
	@echo "   AuthToken string = \"$(authToken)\"" >> src/surf/client/key.go
	@echo "   Subdomain string = \"$(domain)\"" >> src/surf/client/key.go
	@echo ")" >> src/surf/client/key.go

deps: assets
	go get -tags '$(BUILDTAGS)' -d -v surf/...

server: deps
	go install -tags '$(BUILDTAGS)' surf/main/surfd

fmt:
	go fmt surf/...

client: deps
	go install -tags '$(BUILDTAGS)' surf/main/surf

assets: client-assets server-assets

bin/go-bindata:
	GOOS="" GOARCH="" go get github.com/jteeuwen/go-bindata/go-bindata

client-assets: bin/go-bindata
	bin/go-bindata -nomemcopy -pkg=assets -tags=$(BUILDTAGS) \
		-debug=$(if $(findstring debug,$(BUILDTAGS)),true,false) \
		-o=src/surf/client/assets/assets_$(BUILDTAGS).go \
		assets/client/...

server-assets: bin/go-bindata
	bin/go-bindata -nomemcopy -pkg=assets -tags=$(BUILDTAGS) \
		-debug=$(if $(findstring debug,$(BUILDTAGS)),true,false) \
		-o=src/surf/server/assets/assets_$(BUILDTAGS).go \
		assets/server/...

release-client: check-authToken
release-client: check-domain
release-client: BUILDTAGS=release
release-client: client-key
release-client: client

release-server: BUILDTAGS=release
release-server: server

release-all: fmt release-client release-server

all: fmt client server

clean:
	go clean -i -r surf/...
	rm -rf src/surf/client/assets/ src/surf/server/assets/

check-domain:
	ifndef domain
	    $(error domain is required)
	endif

check-authToken:
	ifndef authToken
	    $(error authToken is required)
	endif
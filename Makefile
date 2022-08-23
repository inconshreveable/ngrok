.PHONY: default server client deps clean all release-all contributors

BUILDTAGS=debug
default: all

deps: 
	go mod tidy

server: deps
	go build -tags '$(BUILDTAGS)' -o bin/ngrokd cmd/ngrokd/*.go


client: deps
	go build -tags '$(BUILDTAGS)' -o bin/ngrok cmd/ngrok/*.go

release-client: BUILDTAGS=release
release-client: client

release-server: BUILDTAGS=release
release-server: server

release-all: release-client release-server

all: client server

clean:
	rm -rf bin/{ngrok,ngrokd}

contributors:
	echo "Contributors to ngrok, both large and small:\n" > CONTRIBUTORS
	git log --raw | grep "^Author: " | sort | uniq | cut -d ' ' -f2- | sed 's/^/- /' | cut -d '<' -f1 >> CONTRIBUTORS

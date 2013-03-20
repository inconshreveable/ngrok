.PHONY: default bindir server client

default: client server

bindir:
	mkdir -p bin

server: bindir
	GOPATH=~ go build -o bin/ngrokd main/server.go

client: bindir
	GOPATH=~ go build -o bin/ngrok main/client.go


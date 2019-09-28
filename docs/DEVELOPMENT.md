# Developer's guide to pgrok


## Components
The pgrok project is composed of two components, the pgrok client (pgrok) and the pgrok server (pgrokd).
The pgrok client is the more complicated piece because it has UIs for displaying saved requests and responses.

## Compiling

    git clone git@github.com:jerson/pgrok.git
    cd pgrok && make
    pgrok [LOCAL PORT]


**NB: You must compile with Go 1.13+!.**


## Compiling with docker

```bash
git clone https://github.com/jerson/pgrok && cd pgrok
docker run --rm -it -w /app -v $PWD:/app jerson/go:1.13 sh -c 'make'
```

## Cross compiling with docker

```bash
git clone https://github.com/jerson/pgrok && cd pgrok
docker run --rm -it -w /app -v $PWD:/app jerson/go:1.13 sh -c '
  make deps
  make assets
  mkdir build
  for GOOS in darwin linux windows; do
    for GOARCH in 386 amd64; do
      echo "Building $GOOS-$GOARCH"
      export GOOS=$GOOS
      export GOARCH=$GOARCH
      go build -o ./build/pgrokd-$GOOS-$GOARCH ./cmd/pgrokd
      go build -o ./build/pgrok-$GOOS-$GOARCH ./cmd/pgrok
    done
done
'
```

## Developing locally
The strategy I use for developing on pgrok is to do the following:

Add the following lines to /etc/hosts:

    127.0.0.1 pgrok.me
    127.0.0.1 test.pgrok.me

Run pgrokd with the following options:

    ./bin/pgrokd -domain pgrok.me

Create an pgrok configuration file, "debug.yml" with the following contents:

    server_addr: pgrok.me:4443
    tunnels:
      test:
        proto:
          http: 8080


Then run pgrok with either of these commands:

    ./bin/pgrok -config=debug.yml -log=pgrok.log start test
    ./bin/pgrok -config=debug.yml -log=pgrok.log -subdomain=test 8080

This will get you setup with an pgrok client talking to an pgrok server all locally under your control. Happy hacking!


## Network protocol and tunneling
At a high level, pgrok's tunneling works as follows:

### Connection Setup and Authentication
1. The client initiates a long-lived TCP connection to the server over which they will pass JSON instruction messages. This connection is called the *Control Connection*.
1. After the connection is established, the client sends an *Auth* message with authentication and version information.
1. The server validates the client's *Auth* message and sends an *AuthResp* message indicating either success or failure.

### Tunnel creation
1. The client may then ask the server to create tunnels for it by sending *ReqTunnel* messages. 
1. When the server receives a *ReqTunnel* message, it will send 1 or more *NewTunnel* messages that indicate successful tunnel creation or indicate failure.

### Tunneling connections
1. When the server receives a new public connection, it locates the appropriate tunnel by examining the HTTP host header (or the port number for TCP tunnels). This connection from the public internet is called a *Public Connection*.
1. The server sends a *ReqProxy* message to the client over the control connection.
1. The client initiates a new TCP connection to the server called a *Proxy Connection*.
1. The client sends a *RegProxy* message over the proxy connection so the server can associate it to a control connection (and thus the tunnels it's responsible for).
1. The server sends a *StartProxy* message over the proxy connection with metadata information about the connection (the client IP and name of the tunnel).
1. The server begins copying the traffic byte-for-byte from the public connection to the proxy connection and vice-versa.
1. The client opens a connection to the local address configured for that tunnel. This is called the *Private Connection*.
1. The client begins copying the traffic byte-for-byte from the proxied connection to the private connection and vice-versa.

### Detecting dead tunnels
1. In order to determine whether a tunnel is still alive, the client periodically sends Ping messages over the control connection to the server, which replies with Pong messages.
1. When a tunnel is detected to be dead, the server will clean up all of that tunnel's state and the client will attempt to reconnect and establish a new tunnel.

### Wire format
Messages are sent over the wire as netstrings of the form:

    <message length><message payload>

The message length is sent as a 64-bit little endian integer.

#### msg/msg.go
All of the different message types (Auth, AuthResp, ReqTunnel, RegProxy, StartProxy, etc) are defined here and their fields documented. This is a good place to go to understand exactly what messages are sent between the client and server.
    
## pgrokd - the server
### Code
Code for the server lives under server

### Entry point
The pgrokd entry point is in _/cmd/pgrokd/main.go_.

## pgrok - the client
### Code
Code for the client lives under client

### Entry point
The pgrok entry point is in _cmd/pgrok/main.go_.

## Static assets
The html and javascript code for the pgrok web interface as well as other static assets like TLS/SSL certificates live under the top-level _assets_ directory.

## Beyond
More documentation can be found in the comments of the code itself.

# Developer's guide to ngrok


## Components
The ngrok project is composed of two components, the ngrok client (ngrok) and the ngrok server (ngrokd).
The ngrok client is the more complicated piece because it has UIs for displaying saved requests and responses.

## Compiling

    git clone git@github.com:inconshreveable/ngrok.git
    cd ngrok && make
    bin/ngrok [LOCAL PORT]

There are Makefile targets for compiling just the client or server.

    make client
    make server

**NB: You must compile with Go 1.1+! You must have Mercurial SCM Installed.**

### Compiling release versions
Both the client and the server contain static asset files.
These include TLS/SSL certificates and the html/css/js for the client's web interface.
The release versions embed all of this data into the binaries themselves, whereas the debug versions read these files from the filesystem.

*You should always develop on debug versions so that you don't have to recompile when testing changes in the static assets.*

There are Makefile targets for compiling the client and server for releases:

    make release-client
    make release-server
    make release-all


## Developing locally
The strategy I use for developing on ngrok is to do the following:

Add the following lines to /etc/hosts:

    127.0.0.1 ngrok.me
    127.0.0.1 test.ngrok.me

Run ngrokd with the following options:

    ./bin/ngrokd -domain ngrok.me

Create an ngrok configuration file, "debug.yml" with the following contents:

    server_addr: ngrok.me:4443
    tunnels:
      test:
        proto:
          http: 8080


Then run ngrok with either of these commands:

    ./bin/ngrok -config=debug.yml -log=ngrok.log start test
    ./bin/ngrok -config=debug.yml -log=ngrok.log -subdomain=test 8080

This will get you setup with an ngrok client talking to an ngrok server all locally under your control. Happy hacking!


## Network protocol and tunneling
At a high level, ngrok's tunneling works as follows:

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

### Code
The definitions and shared protocol routines lives under _src/ngrok/msg_

#### src/ngrok/msg/msg.go
All of the different message types (Auth, AuthResp, ReqTunnel, RegProxy, StartProxy, etc) are defined here and their fields documented. This is a good place to go to understand exactly what messages are sent between the client and server.
    
## ngrokd - the server
### Code
Code for the server lives under src/ngrok/server

### Entry point
The ngrokd entry point is in _src/ngrok/server/main.go_.
There is a stub at _src/ngrok/main/ngrokd/ngrokd.go_ for the purposes of creating a properly named binary and being in its own "main" package to comply with go's build system.

## ngrok - the client
### Code
Code for the client lives under src/ngrok/client

### Entry point
The ngrok entry point is in _src/ngrok/client/main.go_.
There is a stub at _src/ngrok/main/ngrok/ngrok.go_ for the purposes of creating a properly named binary and being in its own "main" package to comply with go's build system.

## Static assets
The html and javascript code for the ngrok web interface as well as other static assets like TLS/SSL certificates live under the top-level _assets_ directory.

## Beyond
More documentation can be found in the comments of the code itself.


# pgrok - Introspected tunnels to localhost 
### ”I want to expose a local server behind a NAT or firewall to the internet.”

# Install client

Install supports **Linux** and **MacOS** with **homebrew**
```bash
brew install jerson/tap/pgrok
```

# Install server

Install supports **Linux** and **MacOS** with **homebrew**
```bash
brew install jerson/tap/pgrokd
```

or you can just download it from download section

# Download
just download in [Release section](https://github.com/jerson/pgrok/releases)

# Docker

pgrok and pgrokd available in [Docker Hub](https://hub.docker.com/r/jerson/pgrok)

Sample server in docker-compose

```yaml
version: '3.7'

services:
  pgrokd:
    image: jerson/pgrok
    entrypoint: pgrokd    
    command: -domain sample.com -httpAddr=:80 -httpsAddr=:443 -tunnelAddr=:4443 -tlsCrt=/certs/tls.crt -tlsKey=/certs/tls.key
    volumes:
      - /home/certs:/certs
    ports:
     - 80:80
     - 443:443
     - 4443:4443
```

## What is pgrok?
pgrok is a reverse proxy that creates a secure tunnel from a public endpoint to a locally running web service.
pgrok captures and analyzes all traffic over the tunnel for later inspection and replay.

## What can I do with pgrok?
- Expose any http service behind a NAT or firewall to the internet on a subdomain of pgrok.com
- Expose any tcp service behind a NAT or firewall to the internet on a random port of pgrok.com
- Inspect all http requests/responses that are transmitted over the tunnel
- Replay any request that was transmitted over the tunnel


## What is pgrok useful for?
- Temporarily sharing a website that is only running on your development machine
- Demoing an app at a hackathon without deploying
- Developing any services which consume webhooks (HTTP callbacks) by allowing you to replay those requests
- Debugging and understanding any web service by inspecting the HTTP traffic
- Running networked services on machines that are firewalled off from the internet

## Developing on pgrok
[pgrok developer's guide](docs/DEVELOPMENT.md)

## Compile for linux

```bash
git clone https://github.com/jerson/pgrok && cd pgrok
docker run --rm -it -w /app -v $PWD:/app jerson/go:1.13 sh -c 'make'
```

## Cross compile

```bash
git clone https://github.com/jerson/pgrok && cd pgrok
docker run --rm -it -w /app -v $PWD:/app jerson/go:1.13 sh -c '
  make deps
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


## Disclaimer
pgrok is a fork of ngrok

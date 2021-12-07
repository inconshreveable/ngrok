# pgrok - Introspected tunnels to localhost

### ”I want to expose a local server behind a NAT or firewall to the internet.”

**ejemplo.me** service is shuting down, please take a look, https://github.com/jerson/pgrok/issues/20

# Install client

Install supports **Linux** and **MacOS** with **homebrew**

```bash
brew install jerson/tap/pgrok
```
# Usage

```bash
pgrok -subdomain=customsubdomain 3000
```
sample output

```bash
pgrok                                                           (Ctrl+C to quit)
                                                                                
Tunnel Status                 online                                            
Version                       3.0/3.0                                           
Forwarding                    http://customsubdomain.ejemplo.me -> 127.0.0.1:3000            
Forwarding                    https://customsubdomain.ejemplo.me -> 127.0.0.1:3000           
Web Interface                 http://127.0.0.1:4040                             
# Conn                        0                                                 
Avg Conn Time                 0.00ms 
```

# Downloads

just download in [Release section](https://github.com/jerson/pgrok/releases)

# Install server

Install supports **Linux** and **MacOS** with **homebrew**

```bash
brew install jerson/tap/pgrokd
```

or you can just download it from download section

# Install server with Docker

pgrok and pgrokd available in [Docker Hub](https://hub.docker.com/r/jerson/pgrok)

Sample server in docker-compose

```yaml
version: "3.7"

services:
  pgrokd:
    image: jerson/pgrok
    entrypoint: pgrokd
    command: -domain ejemplo.me -httpAddr=:80 -httpsAddr=:443 -tunnelAddr=:4443 -tlsCrt=/certs/tls.crt -tlsKey=/certs/tls.key
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

- Expose any http service behind a NAT or firewall to the internet on a subdomain of ejemplo.me
- Expose any tcp service behind a NAT or firewall to the internet on a random port of ejemplo.me
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

## Disclaimer

pgrok is a fork of ngrok

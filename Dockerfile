FROM debian:jessie
MAINTAINER Aravind Shankar <aravind@hasura.io>

RUN apt-get update && \
    apt-get install -y build-essential golang git mercurial && \
    mkdir -p /release

COPY device.key /
COPY device.crt /
COPY bin/ngrokd /ngrok/bin/

CMD ["/ngrok/bin/ngrokd", "-tlsKey=/device.key", "-tlsCrt=/device.crt", "-domain=ngrok.hasura.me", "-httpAddr=:80"]

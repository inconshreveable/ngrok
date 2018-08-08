FROM golang:1.10.3-alpine

LABEL maintainer="ferhat.yildiz@turingts.com"

RUN apk add make git

COPY . /tmp

WORKDIR /tmp

RUN rm -rf bin/*

RUN make release-all

RUN mv bin/ngrok* /bin/

ENTRYPOINT ["/bin/ngrok"]

FROM golang:1.16-alpine AS builder

ENV GOPROXY https://proxy.golang.org,direct

ENV WORKDIR /app
WORKDIR ${WORKDIR}

RUN echo "@testing http://nl.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories && \
    echo "@community http://nl.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories

RUN apk add --no-cache binutils make openssh gcc musl-dev build-base \
                       zip unzip git mercurial curl rpm bash jpegoptim \
                       pngquant upx@community

RUN mkdir -p /var/log/go/ && \
    mkdir -p ${GOPATH}/src/ && \
    mkdir -p ${GOPATH}/bin/

ENV PATH ${GOPATH}/bin:/usr/local/go/bin:$PATH

RUN go get -u golang.org/x/lint/golint && \
    go get -u github.com/jteeuwen/go-bindata/go-bindata

COPY go.mod go.sum Makefile ./
RUN make deps

COPY . .
RUN make build

FROM alpine:3.13

LABEL maintainer="jeral17@gmail.com"

RUN apk add --no-cache ca-certificates && update-ca-certificates
RUN apk add --no-cache tzdata

ENV TZ America/Los_Angeles

ENV BUILDER_PATH /app
ENV WORKDIR /app
WORKDIR ${WORKDIR}

COPY --from=builder ${BUILDER_PATH}/pgrok /usr/local/bin/pgrok
COPY --from=builder ${BUILDER_PATH}/pgrokd /usr/local/bin/pgrokd

CMD ["pgrok"]
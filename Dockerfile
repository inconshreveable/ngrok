FROM jerson/go:1.13 AS builder

ENV WORKDIR /app
WORKDIR ${WORKDIR}

COPY go.mod go.sum Makefile ./
RUN make deps

COPY . .
RUN make build

FROM jerson/base:1.2

LABEL maintainer="jeral17@gmail.com"

ENV BUILDER_PATH /app
ENV WORKDIR /app
WORKDIR ${WORKDIR}

COPY --from=builder ${BUILDER_PATH}/pgrok /usr/local/bin/pgrok
COPY --from=builder ${BUILDER_PATH}/pgrokd /usr/local/bin/pgrokd

CMD ["pgrok"]
FROM intersect/debian:latest

COPY entrypoint.sh /entrypoint.sh

COPY bin/ngrokd /usr/local/bin/ngrokd

ENTRYPOINT ["/entrypoint.sh"]

ARG GIT_COMMIT
ARG BUILD_NUMBER

LABEL GIT_COMMIT ${GIT_COMMIT}
LABEL BUILD_NUMBER ${BUILD_NUMBER}

ENV GIT_COMMIT ${GIT_COMMIT}
ENV BUILD_NUMBER ${BUILD_NUMBER}

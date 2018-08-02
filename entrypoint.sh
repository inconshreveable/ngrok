#!/usr/bin/env bash
set -eu

DOMAIN=${DOMAIN:-}
TLS_CERT=${TLS_CERT:-}
TLS_KEY=${TLS_KEY:-}
LOG_LEVEL=${LOG_LEVEL:-debug}

exec /usr/local/bin/ngrokd -log stdout -log-level "$LOG_LEVEL" -d "$DOMAIN" -tlsCrt "$TLS_CERT" -tlsKey "$TLS_KEY"

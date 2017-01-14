#!/usr/bin/env bash
# for Mac 64
make release-all
# for linux 64
GOOS=linux GOARCH=amd64 make release-all
# for windows 64
GOOS=windows GOARCH=amd64 make release-all
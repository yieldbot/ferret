#!/bin/bash

set -ex

GO_LDFLAGS="-X main.version=`git describe --tags --exact-match || git rev-parse --short HEAD` -X main.gitCommit=`git rev-parse --short HEAD`"
BIN_NAME=$(basename `pwd`)
PLATFORMS="darwin linux windows"
rm -rf releases/ && mkdir releases
for PLATFORM in $PLATFORMS; do
  CGO_ENABLED=0 GOOS=${PLATFORM} go build -a -tags netgo -installsuffix netgo -ldflags "$GO_LDFLAGS" -o "$BIN_NAME" && zip "releases/${BIN_NAME}-${PLATFORM}-amd64.zip" "$BIN_NAME" && rm "$BIN_NAME"
done
ls -alF releases/

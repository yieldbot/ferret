#!/bin/bash

set -e

# Run tests
./test.sh

# Initialize required variables
BIN_NAME=$(basename `pwd`)
BIN_VERSION=`git describe --tags --exact-match || git rev-parse --short HEAD || exit 1`
GIT_COMMIT=`git rev-parse --short HEAD || exit 1`
PLATFORMS="darwin linux windows"
ARCHS="amd64"
GO_LDFLAGS="-X main.version=${BIN_VERSION} -X main.commit=${GIT_COMMIT}"
GO_CGO_ENABLED=0

# Prepare release files
mkdir -p releases && rm -rf releases/*.*
echo "release for ${BIN_NAME} ${BIN_VERSION} (${GIT_COMMIT})"
if [ "${BIN_VERSION:0:1}" != "v" ]; then
  echo "no release since current commit ${BIN_VERSION} is not a version tag"
  exit 0
fi

# Iterate platforms
for PLATFORM in $PLATFORMS; do
  BIN_NAME_F=$BIN_NAME
  if [ $PLATFORM == "windows" ]; then
    BIN_NAME_F="${BIN_NAME}.exe"
  fi
  for ARCH in $ARCHS; do
    echo "preparing ${BIN_NAME}-${PLATFORM}-${ARCH}.zip"

    GOOS=${PLATFORM} GOARCH=${ARCH} CGO_ENABLED=${GO_CGO_ENABLED} \
    go build -a -ldflags "$GO_LDFLAGS" -o "${BIN_NAME_F}" && \
    zip "releases/${BIN_NAME}-${PLATFORM}-${ARCH}.zip" "${BIN_NAME_F}" && \
    rm "${BIN_NAME_F}"
  done
done
ls -alF releases/

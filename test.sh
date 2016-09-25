#!/bin/bash

set -ex

# Get required components
go get github.com/golang/lint/golint
go get github.com/rakyll/statik
go generate ./assets/

# Check formatting
OUT=`gofmt -l . | (grep -v '^vendor\/' || true)`; if [ "$OUT" ]; then echo "gofmt: $OUT"; exit 1; fi

# Check linting
OUT=`golint ./... | (grep -v '^vendor\/' || true)`; if [ "$OUT" ]; then echo "golint: $OUT"; exit 1; fi

# Check suspicious constructs
OUT=`find . -type d | grep -v -E '(^.$|^./vendor|^./.git|^./tmp|^./releases|/assets/)' | xargs go vet`; if [ "$OUT" ]; then echo "govet: $OUT"; exit 1; fi

# Build
go build .
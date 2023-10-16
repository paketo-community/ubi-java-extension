#!/usr/bin/env bash

set -euo pipefail

GOOS="linux" go build -ldflags='-s -w' -o bin/helper github.com/paketo-buildpacks/libjvm/v2/cmd/helper
GOOS="linux" go build -ldflags='-s -w' -o bin/main github.com/paketo-community/ubi-java-extension/v1/cmd/main

if [ "${STRIP:-false}" != "false" ]; then
  strip bin/helper bin/main
fi

if [ "${COMPRESS:-none}" != "none" ]; then
  $COMPRESS bin/helper bin/main
fi

ln -fs main bin/generate
ln -fs main bin/detect

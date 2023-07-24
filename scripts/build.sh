#!/usr/bin/env bash

set -euo pipefail

readonly PROGDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly DIR="$(cd "${PROGDIR}/.." && pwd)"

GOOS="linux" go build -ldflags='-s -w' -o bin/helper github.com/paketo-buildpacks/libjvm/v2/cmd/helper
GOOS="linux" go build -ldflags='-s -w' -o bin/main "${DIR}/cmd"

if [ "${STRIP:-false}" != "false" ]; then
  strip bin/helper bin/main
fi

if [ "${COMPRESS:-none}" != "none" ]; then
  $COMPRESS bin/helper bin/main
fi

ln -fs main bin/generate
ln -fs main bin/detect

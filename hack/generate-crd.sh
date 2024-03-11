#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

GOKOALA_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
cd "${GOKOALA_ROOT}"

# Build controller-tools
docker build -f hack/crd/Dockerfile -t gokoala-controller-tools .

# Run against GoKoala
docker run -v `pwd`/:/gokoala gokoala-controller-tools crd paths="/gokoala/hack/crd/..." output:dir="/gokoala/hack/tmp"

# Assertions
cat hack/tmp/pdok_gokoalas.yaml
cat hack/tmp/pdok_gokoalas.yaml | grep "kind: GoKoala"


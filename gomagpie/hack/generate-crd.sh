#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

GOMAGPIE_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
cd "${GOMAGPIE_ROOT}"

if ! command -v controller-gen &> /dev/null
then
    echo "controller-gen not found, using Docker container instead"
    # Build controller-tools
    docker build -f hack/crd/Dockerfile -t gomagpie-controller-tools .

    # Run against GoMagpie
    docker run -v `pwd`/:/gomagpie gomagpie-controller-tools crd paths="/gomagpie/hack/crd/..." output:dir="/gomagpie/hack/tmp"
else
   echo "controller-gen found, using this local install instead of Docker container"

   # Run against GoMagpie config
   controller-gen crd paths="$(pwd)/hack/crd/..." output:dir="$(pwd)/hack/tmp"
fi

# Assertions
cat hack/tmp/pdok_gomagpies.yaml
cat hack/tmp/pdok_gomagpies.yaml | grep "kind: GoMagpie"


#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

GOKOALA_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
cd "${GOKOALA_ROOT}"

if ! command -v controller-gen &> /dev/null
then
    echo "controller-gen not found, using Docker container instead"
    # Build controller-tools
    docker build -f hack/crd/Dockerfile -t gokoala-controller-tools .

    # Run against GoKoala
    docker run -v `pwd`/:/gokoala gokoala-controller-tools crd paths="/gokoala/hack/crd/..." output:dir="/gokoala/hack/tmp"
else
   echo "controller-gen found, using this local install instead of Docker container"

   # Run against GoKoala config
   controller-gen object paths="$(pwd)/hack/crd/..." output:dir="$(pwd)hack/tmp"
fi

# Assertions
cat hack/tmp/pdok_gokoalas.yaml
cat hack/tmp/pdok_gokoalas.yaml | grep "kind: GoKoala"


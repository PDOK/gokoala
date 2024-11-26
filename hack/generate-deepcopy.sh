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

    # Run against GoMagpie config
    docker run -v `pwd`/:/gomagpie gomagpie-controller-tools object paths="/gomagpie/config/..." output:dir="/gomagpie/config/"
else
   echo "controller-gen found, using this local install instead of Docker container"

   # Run against GoMagpie config
   controller-gen object paths="$(pwd)/config/..." output:dir="$(pwd)/config/"
fi

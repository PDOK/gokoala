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

    # Run against GoKoala config
    docker run -v `pwd`/:/gokoala gokoala-controller-tools object paths="/gokoala/config/..." output:dir="/gokoala/config/"
else
   echo "controller-gen found, using this local install instead of Docker container"

   # Run against GoKoala config
   controller-gen object paths="$(pwd)/config/..." output:dir="$(pwd)/config/"
fi

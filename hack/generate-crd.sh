#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

GOKOALA_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
cd "${GOKOALA_ROOT}"

docker build -f hack/crd/Dockerfile -t gokoala-controller-tools .
docker run -v `pwd`/:/gokoala gokoala-controller-tools crd paths="/gokoala/hack/crd/..." output:stdout


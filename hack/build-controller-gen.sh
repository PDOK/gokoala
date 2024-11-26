#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

git clone https://github.com/kubernetes-sigs/controller-tools.git ct && \
    cd ct && \
    cd cmd/controller-gen && \
    go build -o /bin/controller-gen . && \
    /bin/controller-gen --help
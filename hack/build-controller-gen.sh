#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

# Note we build from PR https://github.com/kubernetes-sigs/controller-tools/pull/892 until its merged to master
git clone https://github.com/kubernetes-sigs/controller-tools.git ct && \
    cd ct && \
    git fetch origin pull/892/head:pull892 && \
    git checkout pull892 && \
    cd cmd/controller-gen && \
    go build -o /bin/controller-gen . && \
    /bin/controller-gen --help
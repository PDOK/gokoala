#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.19.0 && \
    controller-gen --help
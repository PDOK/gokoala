#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

GOKOALA_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
cd "${GOKOALA_ROOT}"

if ! command -v type2md &> /dev/null
then
  echo "installing type2md"
  # using fork: https://github.com/rkettelerij/type2md/tree/fork maybe permanently but at least
  # until PRs https://github.com/eleztian/type2md/pull/1 and https://github.com/eleztian/type2md/pull/2 are merged
  go get github.com/rkettelerij/type2md@c111a3b53690f56f658c742bda79448ed0b398e6
  go install github.com/rkettelerij/type2md@c111a3b53690f56f658c742bda79448ed0b398e6
fi

echo "generating markdown configuration reference from config structs"
type2md -t "Configuration Reference" -f docs/README.md github.com/PDOK/gokoala/config Config
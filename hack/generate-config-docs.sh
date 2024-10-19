#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

GOKOALA_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
cd "${GOKOALA_ROOT}"

if ! command -v type2md &> /dev/null
then
  echo "installing type2md tool"
  # using fork: https://github.com/rkettelerij/type2md/tree/fork, likely permanently but at least
  # until the following PRs are merged.
  # - https://github.com/eleztian/type2md/pull/1
  # - https://github.com/eleztian/type2md/pull/2
  go install -mod readonly github.com/rkettelerij/type2md@c111a3b53690f56f658c742bda79448ed0b398e6
fi

echo "generating markdown configuration reference from config structs"
type2md -t "Configuration Reference" -f docs/README.md github.com/PDOK/gokoala/config Config
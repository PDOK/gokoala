#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

GOKOALA_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
cd "${GOKOALA_ROOT}"

echo "npm install"
npm install --prefix viewer

echo "angular build"
npm run build --prefix viewer

echo "place viewer in assets"
cp -r viewer/dist/view-component assets/

echo "done, now start GoKoala (go run main.go) and checkout the vectortile viewer"

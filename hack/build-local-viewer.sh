#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

GOKOALA_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
cd "${GOKOALA_ROOT}"

echo "npm install"
npm install --force --prefix viewer

echo "angular build"
npm run build --prefix viewer

echo "place viewer in assets"
rm -rf assets/view-component
cp -r viewer/dist/view-component/browser assets/view-component

echo "done, now start GoKoala (go run main.go) and checkout the vectortile viewer"

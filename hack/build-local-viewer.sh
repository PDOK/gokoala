#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

GOKOALA_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
cd "${GOKOALA_ROOT}"

echo "npm install"
npm install --prefix webcomponents/vectortile-view-component

echo "angular build"
npm run build --prefix webcomponents/vectortile-view-component

echo "place viewer in assets"
cp -r webcomponents/vectortile-view-component/dist/vectortile-view-component assets/

echo "done, now start GoKoala (go run main.go) and checkout the vectortile viewer"

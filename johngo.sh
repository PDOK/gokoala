#!/bin/sh
./hack/build-local-viewer.sh
go run main.go --config-file ./examples/config_features_local.yaml
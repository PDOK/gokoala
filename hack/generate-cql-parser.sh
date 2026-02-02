#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

GOKOALA_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
cd "${GOKOALA_ROOT}"

# ANTLR version used, should match version of github.com/antlr4-go/antlr in go.mod
ANTLR_VERSION=4.13.1

ANTLR_PARAMS="-Dlanguage=Go -no-visitor -package parser CqlLexer.g4 CqlParser.g4"

# store output of `java -version`
java_version_raw=$(java -version 2>&1 | head -n 1)
# extract major version number (first integer).
java_major=$(echo "$java_version_raw" | grep -oE '[0-9]+' | head -n 1)

if [[ "$java_major" =~ ^[0-9]+$ ]] && (( java_major >= 17 )); then
  echo "Java 17 or newer detected (version $java_major). Using ANTLR directly without Docker"

  # Run against CQL grammar (note: when updating this command also change it below for the Docker variant)
  cd internal/ogc/features/cql/parser
  java -Xmx256M -jar ../../../../../hack/antlr/antlr-${ANTLR_VERSION}-complete.jar ${ANTLR_PARAMS}
else
  echo "Java 17+ not found, using Docker to run ANTLR"

  # Build ANTLR container
  echo "build ANTLR container"
  docker build --build-arg ANTLR_VERSION=${ANTLR_VERSION} -f hack/antlr/Dockerfile -t antlr:local hack/antlr

  # Run against CQL grammar (note: when updating this command also change it above for the plain Java variant)
  echo "running ANTLR to generate CQL parser"
  cd internal/ogc/features/cql/parser
  docker run --rm -v `pwd`/:/work -w /work antlr:local ${ANTLR_PARAMS}
fi
echo "finished generating CQL parser"

printf "formatting Go code including generated code\n\n"
go fmt ./..
printf "DONE!\n"

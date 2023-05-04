ARG REGISTRY="docker.io"
FROM ${REGISTRY}/golang:1.20 AS build-env

WORKDIR /go/src/service
ADD . /go/src/service

# disable crosscompiling
ENV CGO_ENABLED=0
# compile linux only
ENV GOOS=linux

RUN go mod download all

# build the binary with debug information removed.
# also run tests, the short flag skips integration tests since we can't run Testcontainers in multistage Docker :-(
RUN go test -short && go build -v -ldflags '-w -s' -a -installsuffix cgo -o /gokoala github.com/PDOK/gokoala

# delete all go files (and testdata dirs) so only assets/templates/etc remain, since in a later
# stage we need to copy these remaining files including their subdirectories to the final docker image.
RUN find . -type f -name "*.go" -delete && find . -type d -name "testdata" -prune -exec rm -rf {} \;

##########################################################
FROM scratch
EXPOSE 8080
# use the WORKDIR to create a /tmp folder, mkdir is not available
WORKDIR /tmp
WORKDIR /
ENV PATH=/

# include executable
COPY --from=build-env /gokoala /

# include assets/templates/etc (be specific here to only include required dirs)
COPY --from=build-env /go/src/service/assets/ /assets/
COPY --from=build-env /go/src/service/engine/ /engine/
COPY --from=build-env /go/src/service/ogc/ /ogc/

COPY --from=build-env /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build-env /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# run as non-root
USER 1001
ENTRYPOINT ["gokoala"]

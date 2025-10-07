####### Go build
FROM docker.io/golang:1.25-bookworm AS build-env
WORKDIR /go/src/service
ADD . /go/src/service

# enable cgo in order to interface with sqlite
ENV CGO_ENABLED=1
ENV GOOS=linux

# install sqlite-related compile-time dependencies
RUN set -eux && \
    apt-get update && \
    apt-get install -y libsqlite3-mod-spatialite && \
    rm -rf /var/lib/apt/lists/*

# build & test the binary with debug information removed.
RUN go mod download all && \
    go test -short ./... && \
    go build -v -ldflags '-w -s' -a -installsuffix cgo -o /gomagpie github.com/PDOK/gomagpie/cmd

# delete all go files (and testdata dirs) so only assets/templates/etc remain, since in a later
# stage we need to copy these remaining files including their subdirectories to the final docker image.
RUN find . -type f -name "*.go" -delete && find . -type d -name "testdata" -prune -exec rm -rf {} \;

####### Final image (use debian tag since we rely on C-libs)
FROM docker.io/debian:bookworm-slim

EXPOSE 8080
# use the WORKDIR to create a /tmp folder
WORKDIR /tmp
WORKDIR /

# install sqlite-related runtime dependencies
RUN set -eux && \
    apt-get update && \
    apt-get install -y libsqlite3-mod-spatialite && \
    rm -rf /var/lib/apt/lists/*

# include executable
COPY --from=build-env /gomagpie /

# include assets/templates/etc (be specific here to only include required dirs)
COPY --from=build-env /go/src/service/assets/ /assets/
COPY --from=build-env /go/src/service/internal/ /internal/

# run as non-root
USER 1001
ENTRYPOINT ["/gomagpie"]

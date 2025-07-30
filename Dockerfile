####### Node.js build
FROM docker.io/node:lts-alpine3.17 AS build-component
RUN mkdir -p /usr/src/app
COPY ./viewer /usr/src/app
WORKDIR /usr/src/app
RUN npm config set registry http://registry.npmjs.org && \
    npm install && \
    npm run build

####### Go build
FROM docker.io/golang:1.24-bookworm AS build-env
WORKDIR /go/src/service
COPY . /go/src/service

# enable cgo in order to interface with sqlite
ENV CGO_ENABLED=1
ENV GOOS=linux

# install sqlite-related compile-time dependencies
RUN set -eux && \
    apt-get update && \
    apt-get install --no-install-recommends -y  \
      libcurl4-openssl-dev=*  \
      libssl-dev=*  \
      libsqlite3-mod-spatialite=* && \
    rm -rf /var/lib/apt/lists/*

# install controller-gen (used by go generate)
RUN hack/build-controller-gen.sh

# build the binary with debug information removed.
# no tests are run since some tests rely on Testcontainers which doesn't work in multi-stage builds (dind).
RUN go mod download all && \
    go generate -v ./... && \
    go build -v -ldflags '-w -s' -a -installsuffix cgo -o /gokoala github.com/PDOK/gokoala/cmd

# delete all go files (and testdata dirs) so only assets/templates/etc remain, since in a later
# stage we need to copy these remaining files including their subdirectories to the final docker image.
RUN find . -type f -name "*.go" -delete && find . -type d -name "testdata" -prune -exec rm -rf {} \;

####### Final image (use debian tag since we rely on C-libs)
FROM docker.io/debian:bookworm-slim

# install sqlite-related runtime dependencies (also make sure to add these to GH actions workflows)
RUN set -eux && \
    apt-get update && \
    apt-get install --no-install-recommends -y  \
      libcurl4=*  \
      curl=*  \
      openssl=*  \
      ca-certificates=*  \
      libsqlite3-mod-spatialite=* \
      proj-bin=* && \
    rm -rf /var/lib/apt/lists/*

EXPOSE 8080
# use the WORKDIR to create a /tmp folder
WORKDIR /tmp
WORKDIR /

# include executable
COPY --from=build-env /gokoala /

# include assets/templates/etc (be specific here to only include required dirs)
COPY --from=build-env /go/src/service/assets/ /assets/
COPY --from=build-env /go/src/service/internal/ /internal/
COPY --from=build-env /go/src/service/themes/ /themes

# include viewer as asset
COPY --from=build-component /usr/src/app/dist/view-component/browser/*.js  /assets/view-component/
COPY --from=build-component /usr/src/app/dist/view-component/browser/*.css  /assets/view-component/
COPY --from=build-component /usr/src/app/dist/view-component/3rdpartylicenses.txt /assets/view-component/3rdpartylicenses.txt

# run as non-root
USER 1001
ENTRYPOINT ["/gokoala"]

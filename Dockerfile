ARG REGISTRY="docker.io"

####### Node.js build
FROM ${REGISTRY}/node:lts-alpine3.17 AS build-component
RUN mkdir -p /usr/src/app
COPY ./webcomponents/vectortile-view-component /usr/src/app
WORKDIR /usr/src/app
RUN npm install
RUN npm run build

####### Go build
FROM ${REGISTRY}/golang:1.21-bookworm AS build-env
WORKDIR /go/src/service
ADD . /go/src/service

# enable cgo in order to interface with sqlite
ENV CGO_ENABLED=1
ENV GOOS=linux

# install sqlite-related compile-time dependencies
RUN set -eux && \
    apt-get update && \
    apt-get install -y libcurl4-openssl-dev libssl-dev libsqlite3-mod-spatialite && \
    rm -rf /var/lib/apt/lists/*

RUN go mod download all

# build & test the binary with debug information removed.
RUN go test -short && \
    go build -v -ldflags '-w -s' -a -installsuffix cgo -o /gokoala github.com/PDOK/gokoala

# delete all go files (and testdata dirs) so only assets/templates/etc remain, since in a later
# stage we need to copy these remaining files including their subdirectories to the final docker image.
RUN find . -type f -name "*.go" -delete && find . -type d -name "testdata" -prune -exec rm -rf {} \;

####### Final image (use debian tag since we rely on C-libs)
FROM ${REGISTRY}/debian:bookworm-slim

# install sqlite-related runtime dependencies
RUN set -eux && \
    apt-get update && \
    apt-get install -y libcurl4 openssl libsqlite3-mod-spatialite && \
    rm -rf /var/lib/apt/lists/*

EXPOSE 8080
# use the WORKDIR to create a /tmp folder
WORKDIR /tmp
WORKDIR /
ENV PATH=/

# include executable
COPY --from=build-env /gokoala /

# include assets/templates/etc (be specific here to only include required dirs)
COPY --from=build-env /go/src/service/assets/ /assets/
# include vectortile-view-component webcomponent as asset
COPY --from=build-component /usr/src/app/dist/vectortile-view-component/styles.css  /assets/vectortile-view-component/styles.css
COPY --from=build-component /usr/src/app/dist/vectortile-view-component/main.js  /assets/vectortile-view-component/main.js
COPY --from=build-component /usr/src/app/dist/vectortile-view-component/polyfills.js  /assets/vectortile-view-component/polyfills.js
COPY --from=build-component /usr/src/app/dist/vectortile-view-component/runtime.js  /assets/vectortile-view-component/runtime.js
COPY --from=build-component /usr/src/app/dist/vectortile-view-component/3rdpartylicenses.txt /assets/vectortile-view-component/3rdpartylicenses.txt

COPY --from=build-env /go/src/service/engine/ /engine/
COPY --from=build-env /go/src/service/ogc/ /ogc/

# run as non-root
USER 1001
ENTRYPOINT ["gokoala"]

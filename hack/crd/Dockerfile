FROM docker.io/golang:1.22-bookworm AS build-env
ADD hack/build-controller-gen.sh /build-controller-gen.sh
RUN /build-controller-gen.sh

FROM docker.io/golang:1.22-bookworm
COPY --from=build-env /bin/controller-gen /bin/controller-gen
ENTRYPOINT ["/bin/controller-gen"]
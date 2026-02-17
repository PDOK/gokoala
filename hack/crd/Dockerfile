FROM docker.io/golang:1.26-bookworm AS build-env
RUN go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.19.0
ENTRYPOINT ["controller-gen"]
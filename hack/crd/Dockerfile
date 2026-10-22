FROM docker.io/golang:1.26-bookworm AS build-env
RUN apt-get update && \
    apt-get install --no-install-recommends -y pkg-config libproj-dev && \
    rm -rf /var/lib/apt/lists/*
RUN go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.19.0
ENTRYPOINT ["controller-gen"]
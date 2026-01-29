FROM eclipse-temurin:25-jre

# ANTLR version used, should match version of github.com/antlr4-go/antlr in go.mod
# Tip: grab new versions from https://www.antlr.org/download/antlr-${ANTLR_VERSION}-complete.jar.
ARG ANTLR_VERSION

# Install ANTLR
COPY antlr-${ANTLR_VERSION}-complete.jar /usr/local/lib/antlr.jar

# Add ANTLR to PATH via a wrapper script
RUN echo '#!/bin/sh\njava -Xmx256M -jar /usr/local/lib/antlr.jar "$@"' \
    > /usr/local/bin/antlr && chmod +x /usr/local/bin/antlr

WORKDIR /work

ENTRYPOINT ["antlr"]

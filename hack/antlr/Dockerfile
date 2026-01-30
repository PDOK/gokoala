FROM eclipse-temurin:25-jre

# ANTLR version used, should match version of github.com/antlr4-go/antlr in go.mod
ARG ANTLR_VERSION

# Install ANTLR. New versions can be downloadend from https://www.antlr.org/download/antlr-${ANTLR_VERSION}-complete.jar.
COPY antlr-${ANTLR_VERSION}-complete.jar /usr/local/lib/antlr.jar

WORKDIR /work

ENTRYPOINT ["java", "-Xmx256M", "-jar", "/usr/local/lib/antlr.jar"]

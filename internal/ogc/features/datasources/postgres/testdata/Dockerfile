# GDAL with Postgres client
FROM ghcr.io/osgeo/gdal:ubuntu-small-3.11.3

RUN apt-get update && \
    apt-get install -y postgresql-client && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

RUN ogrinfo --version && psql --version

CMD ["/bin/bash", "-l"]
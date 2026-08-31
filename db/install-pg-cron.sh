#!/bin/sh

set -eu

PG_CRON_VERSION="${PG_CRON_VERSION:-1.6.7}"
PG_CONFIG="/usr/local/bin/pg_config"

PG_CRON_LIB="$($PG_CONFIG --pkglibdir)/pg_cron.so"
PG_CRON_CONTROL="$($PG_CONFIG --sharedir)/extension/pg_cron.control"

echo "Checking pg_cron..."

if [ ! -f "$PG_CRON_LIB" ] || [ ! -f "$PG_CRON_CONTROL" ]; then

    echo "Installing pg_cron ${PG_CRON_VERSION}..."

    apk add --no-cache --virtual .pgcron-build-deps \
        build-base \
        curl

    BUILD_DIR="$(mktemp -d)"

    curl -fsSL \
        "https://github.com/citusdata/pg_cron/archive/refs/tags/v${PG_CRON_VERSION}.tar.gz" \
        -o "${BUILD_DIR}/pg_cron.tar.gz"

    tar -xzf "${BUILD_DIR}/pg_cron.tar.gz" \
        -C "${BUILD_DIR}"

    cd "${BUILD_DIR}/pg_cron-${PG_CRON_VERSION}"

    make PG_CONFIG="$PG_CONFIG"

    make \
        PG_CONFIG="$PG_CONFIG" \
        install

    cd /

    rm -rf "$BUILD_DIR"

    apk del .pgcron-build-deps

    echo "pg_cron ${PG_CRON_VERSION} installed successfully."

else

    echo "pg_cron already installed."

fi


# Continue with the original PostgreSQL entrypoint.
exec /usr/local/bin/docker-entrypoint.sh "$@"
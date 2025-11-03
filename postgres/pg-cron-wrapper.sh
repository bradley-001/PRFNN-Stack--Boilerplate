#!/bin/bash
set -e

# Set the pg-cron database name from POSTGRES_DB environment variable
if [ -n "$POSTGRES_DB" ]; then
    echo "cron.database_name = '$POSTGRES_DB'" >> /usr/share/postgresql/postgresql.conf.sample
fi

# Call the original entrypoint
exec docker-entrypoint.sh "$@"
FROM postgres:18

RUN apt-get update \
    && apt-get install -y postgresql-18-cron \
    && rm -rf /var/lib/apt/lists/*

RUN echo "shared_preload_libraries = 'pg_cron'" >> /usr/share/postgresql/postgresql.conf.sample

# Custom entrypoint wrapper
COPY pg-cron-wrapper.sh /usr/local/bin
RUN chmod +x /usr/local/bin/pg-cron-wrapper.sh

# SQL Init script
COPY init.sql /docker-entrypoint-initdb.d/01-init.sql

ENTRYPOINT [ "pg-cron-wrapper.sh" ]
CMD ["postgres"]
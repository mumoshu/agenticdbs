FROM postgres:16

# Build dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    git \
    postgresql-server-dev-16 \
    cmake \
    flex \
    bison \
    libreadline-dev \
    curl \
    ca-certificates \
    apt-transport-https \
    lsb-release

# pgvector — vector similarity search
RUN cd /tmp \
    && git clone --branch v0.8.0 --depth 1 https://github.com/pgvector/pgvector.git \
    && cd pgvector \
    && make \
    && make install \
    && rm -rf /tmp/pgvector

# Apache AGE — graph database
RUN cd /tmp \
    && git clone --branch PG16/v1.5.0-rc0 --depth 1 https://github.com/apache/age.git \
    && cd age \
    && make \
    && make install \
    && rm -rf /tmp/age

# TimescaleDB — timeseries
RUN curl -s https://packagecloud.io/install/repositories/timescale/timescaledb/script.deb.sh | bash \
    && apt-get install -y timescaledb-2-postgresql-16

# Configure shared_preload_libraries for extensions that need it
RUN echo "shared_preload_libraries = 'age,timescaledb'" >> /usr/share/postgresql/postgresql.conf.sample

# Cleanup build dependencies to reduce image size
RUN apt-get purge -y --auto-remove \
    build-essential git cmake flex bison libreadline-dev \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

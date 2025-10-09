# The dockerfile is currently still WIP and might be broken
FROM golang:1.23.5 AS build-env

# Install latest Node.js (24.x) and npm for bundling API docs
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    ca-certificates \
    gnupg \
  && mkdir -p /etc/apt/keyrings \
  && curl -fsSL https://deb.nodesource.com/gpgkey/nodesource-repo.gpg.key | gpg --dearmor -o /etc/apt/keyrings/nodesource.gpg \
  && echo "deb [signed-by=/etc/apt/keyrings/nodesource.gpg] https://deb.nodesource.com/node_24.x nodistro main" > /etc/apt/sources.list.d/nodesource.list \
  && apt-get update \
  && apt-get install -y --no-install-recommends nodejs \
  && apt-get purge -y gnupg \
  && rm -rf /var/lib/apt/lists/*

WORKDIR /src

# Install JS deps used for API docs bundling
COPY package.json package-lock.json ./
RUN npm ci

# Prepare Go dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source and build
ADD . ./
ARG target=all
RUN make -B $target

# final stage
FROM ubuntu:22.04
RUN apt-get update && apt-get -y upgrade && apt-get install -y --no-install-recommends \
  libssl-dev \
  ca-certificates \
  && apt-get clean \
  && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=build-env /src/bin /app/
COPY --from=build-env /src/config /app/config
CMD ["./explorer", "--config", "./config/default.config.yml"]
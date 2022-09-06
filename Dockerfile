# The dockerfile is currently still WIP and might be broken
FROM golang:1.18 AS build-env
ADD . /src
WORKDIR /src
RUN go mod download
RUN make -B all

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
# The dockerfile is currently still WIP and might be broken
FROM golang:1.20.8 AS build-env
COPY go.mod go.sum /src/
WORKDIR /src
RUN go mod download
RUN go install github.com/swaggo/swag/cmd/swag@v1.8.3
ADD . /src
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
# The dockerfile is currently still WIP and might be broken
FROM golang:1.17-alpine AS build-env
RUN apk --no-cache add build-base git musl-dev linux-headers npm
ADD . /src
WORKDIR /src
RUN make -B all

# final stage
FROM alpine
WORKDIR /app
RUN apk --no-cache add libstdc++ libgcc
COPY --from=build-env /src/bin /app/
COPY --from=build-env /src/phase0.yml /app/phase0.yml
COPY  ./config-example.yml /app/config.yml
CMD ["./explorer", "--config", "config.yml"]

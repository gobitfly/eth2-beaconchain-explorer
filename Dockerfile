# The dockerfile is currently still WIP and might be broken
FROM golang:alpine AS build-env
RUN apk --no-cache add build-base git bzr mercurial gcc musl-dev linux-headers npm
ADD . /src
RUN cd /src && make -B all

# final stage
FROM alpine
WORKDIR /app
COPY --from=build-env /src/bin /app/
COPY ./config-example.yml /app/config.yml
CMD ["./explorer"]


# # The dockerfile is currently still WIP and might be broken
# FROM golang:1.14 AS build-env
# RUN apt-get update && apt-get install -y git bzr mercurial gcc nodejs npm
# ADD . /src
# RUN cd /src && make -B all
# 
# # final stage
# FROM ubuntu:18.04
# WORKDIR /app
# COPY --from=build-env /src/bin /app/
# COPY ./config-example.yml /app/config.yml
# CMD ["./explorer"]

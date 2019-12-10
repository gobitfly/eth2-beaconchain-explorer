# The dockerfile is currently still WIP and might be broken
FROM golang:alpine AS build-env
RUN apk --no-cache add build-base git bzr mercurial gcc
ADD . /src
RUN cd /src && make all

# final stage
FROM alpine
WORKDIR /app
COPY --from=build-env /src/bin /app/
COPY ./config-stefan.yml /app/config.yml
CMD ["./explorer"]
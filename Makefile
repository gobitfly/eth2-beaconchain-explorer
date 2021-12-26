GITCOMMIT=`git describe --always`
VERSION=$$(git describe 2>/dev/null || echo "0.0.0-${GITCOMMIT}")
GITDATE=`TZ=UTC git show -s --date=iso-strict-local --format=%cd HEAD`
BUILDDATE=`date -u +"%Y-%m-%dT%H:%M:%S%:z"`
PACKAGE=eth2-exporter
LDFLAGS="-X ${PACKAGE}/version.Version=${VERSION} -X ${PACKAGE}/version.BuildDate=${BUILDDATE} -X ${PACKAGE}/version.GitCommit=${GITCOMMIT} -X ${PACKAGE}/version.GitDate=${GITDATE}"

all: explorer stats

lint:
	golint ./...

test:
	go test -tags=blst_enabled ./...

explorer:
	rm -rf bin/
	mkdir -p bin/
	cp -r templates bin/
	go run cmd/bundle/main.go
	cp -r static/ bin/static
	cp -r locales/ bin/
	cp -r config/ bin/config
	go get github.com/swaggo/swag/cmd/swag@v1.7.4 && swag init -g handlers/api.go
	go build --ldflags=${LDFLAGS} -o bin/explorer cmd/explorer/main.go

chartshotter:
	go build --ldflags=${LDFLAGS} -o bin/chartshotter cmd/chartshotter/main.go

stats:
	go build --ldflags=${LDFLAGS} -o bin/statistics cmd/statistics/main.go
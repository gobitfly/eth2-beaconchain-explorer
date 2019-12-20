GITCOMMIT=`git describe --always`
VERSION=$$(git describe 2>/dev/null || echo "0.0.0-${GITCOMMIT}")
GITDATE=`TZ=UTC git show -s --date=iso-strict-local --format=%cd HEAD`
BUILDDATE=`date -u +"%Y-%m-%dT%H:%M:%S%:z"`
PACKAGE=eth2-exporter
LDFLAGS="-X ${PACKAGE}/version.Version=${VERSION} -X ${PACKAGE}/version.BuildDate=${BUILDDATE} -X ${PACKAGE}/version.GitCommit=${GITCOMMIT} -X ${PACKAGE}/version.GitDate=${GITDATE}"

all: bootstrap explorer

lint:
	golint ./...

bootstrap:
	npm ci --prefix ./bootstrap && npm run --prefix ./bootstrap dist-css

explorer:
	rm -rf bin/
	mkdir -p bin/templates/
	cp -r templates/ bin/
	cp -r static/ bin/static
	go build --ldflags=${LDFLAGS} -o bin/explorer cmd/explorer/main.go


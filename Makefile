GITCOMMIT=`git describe --always`
VERSION=$$(git describe 2>/dev/null || echo "0.0.0-${GITCOMMIT}")
GITDATE=`TZ=UTC git show -s --date=iso-strict-local --format=%cd HEAD`
BUILDDATE=`date -u +"%Y-%m-%dT%H:%M:%S%:z"`
PACKAGE=eth2-exporter
LDFLAGS="-X ${PACKAGE}/version.Version=${VERSION} -X ${PACKAGE}/version.BuildDate=${BUILDDATE} -X ${PACKAGE}/version.GitCommit=${GITCOMMIT} -X ${PACKAGE}/version.GitDate=${GITDATE}"
ETH2APIREMOTE=https://github.com/ethereum/eth2.0-APIs.git
ETH2APITMPDIR=/tmp/eth2api

all: explorer

lint:
	golint ./...

explorer:
	rm -rf bin/
	mkdir -p bin/templates/
	cp -r templates/ bin/
	cp -r static/ bin/static
	go build --ldflags=${LDFLAGS} -o bin/explorer cmd/explorer/main.go
	go build --ldflags=${LDFLAGS} -o bin/chartshotter cmd/chartshotter/main.go

eth2api2:
	rm -rf $(ETH2APITMPDIR)
	git clone $(ETH2APIREMOTE) $(ETH2APITMPDIR)
	# mkdir -p $(ETH2APITMPDIR)
	# wget -q -O $(ETH2APITMPDIR)/beacon-node-oapi.yaml https://ethereum.github.io/eth2.0-APIs/beacon-node-oapi.yaml
	ls $(ETH2APITMPDIR)
	docker run --rm \
		-v $(ETH2APITMPDIR):/v \
		guybrush/swagger-cli swagger-cli bundle -r /v/beacon-node-oapi.yaml > $(ETH2APITMPDIR)/bundle.yaml
	rm -rf eth2api2
	docker run --rm \
		-v $(CURDIR):/v \
		-v $(ETH2APITMPDIR):/v2 \
		swaggerapi/swagger-codegen-cli generate \
		-i /v2/bundle.yaml \
		-l go \
		-DpackageName=eth2api2 \
		-o /v/eth2api2
	sudo chown -R 1000:1000 eth2api2
	rm -rf eth2api2/go.{mod,sum}
	# rm -rf eth2api2/api_validator_required_api.go
	go build eth2api2/*.go

# eth2api2:
# 	rm -rf $(ETH2APITMPDIR)
# 	# git clone $(ETH2APIREMOTE) $(ETH2APITMPDIR)
# 	mkdir -p $(ETH2APITMPDIR)
# 	wget -q -O $(ETH2APITMPDIR)/beacon-node-oapi.yaml https://ethereum.github.io/eth2.0-APIs/beacon-node-oapi.yaml
# 	ls $(ETH2APITMPDIR)
# 	docker run --rm \
# 		-v $(ETH2APITMPDIR):/v \
# 		guybrush/swagger-cli swagger-cli bundle -r /v/beacon-node-oapi.yaml > $(ETH2APITMPDIR)/bundle.yaml
# 	rm -rf eth2api2
# 	docker run --rm \
# 		-v $(CURDIR):/v \
# 		-v $(ETH2APITMPDIR):/v2 \
# 		openapitools/openapi-generator-cli generate \
# 		-i /v2/bundle.yaml \
# 		-g go \
# 		--additional-properties=packageName=eth2api2,isGoSubmodule=true \
# 		-o /v/eth2api2
# 	sudo chown -R 1000:1000 eth2api2
# 	rm -rf eth2api2/go.{mod,sum}
# 	rm -rf eth2api2/api_validator_required_api.go
# 	go build eth2api2/*.go

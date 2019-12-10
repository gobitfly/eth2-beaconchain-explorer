all: explorer

explorer:
	rm -rf bin/
	mkdir -p bin/templates/
	cp -r templates/ bin/
	cp -r static/ bin/static
	go build -o bin/explorer cmd/explorer/main.go

run:
	go run -race cmd/explorer/main.go --config "config-stefan.yml"
